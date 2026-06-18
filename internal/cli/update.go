package cli

import (
	"archive/tar"
	"bufio"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
	"golang.org/x/mod/semver"
)

const releaseBaseURL = "https://tene-releases.s3.ap-northeast-2.amazonaws.com"

var updateCmd = &cobra.Command{
	Use:   "update [version]",
	Short: "Update tene to the latest or specified version",
	Long: `Update tene to the latest version or a specific version.

Examples:
  tene update            # Update to latest version
  tene update v0.2.0     # Update to specific version
  tene update --check    # Check for updates without installing`,
	Args: cobra.MaximumNArgs(1),
	RunE: runUpdate,
}

var (
	updateFlagCheck             bool
	updateFlagIncludePrerelease bool
)

func init() {
	updateCmd.Flags().BoolVar(&updateFlagCheck, "check", false, "Check for updates without installing")
	updateCmd.Flags().BoolVar(&updateFlagIncludePrerelease, "include-prerelease", false,
		"Allow upgrading to RC/beta releases (opt-in)")
}

// shouldOfferUpdate reports whether tene should offer to upgrade from
// `current` to `latest`. Sprint v1014-rc1-qa-fixes / FX3 (invariant I-13).
//
// The rules are deliberately conservative — the worst-case behaviour of
// rc1's naive != comparison was to recommend a downgrade from v1.0.14-rc1
// to v1.0.13. Every clause below corresponds to a documented scenario in
// internal/cli/update_semver_test.go.
//
//   - current is "dev"/"vdev"/empty: no comparable baseline; never offer.
//   - either tag is not valid semver: malformed input; fail closed.
//   - semver.Compare(latest, current) <= 0: latest is not strictly newer;
//     covers both "already up to date" and the B3 downgrade case.
//   - latest is a pre-release and the user did not opt in via
//     --include-prerelease: never auto-recommend RC/beta to stable users.
//   - otherwise: offer.
//
// An explicit `tene update v1.0.13` (user supplies a target version) does
// NOT pass through this helper — that path is a manual downgrade, which
// is left available because it is sometimes legitimate (rolling back a
// broken release).
func shouldOfferUpdate(current, latest string, includePrerelease bool) bool {
	if current == "vdev" || current == "dev" || current == "" {
		return false
	}
	if !semver.IsValid(current) || !semver.IsValid(latest) {
		return false
	}
	if semver.Compare(latest, current) <= 0 {
		return false
	}
	if !includePrerelease && semver.Prerelease(latest) != "" {
		return false
	}
	return true
}

type releaseInfo struct {
	Version     string
	DownloadURL string
	HTMLURL     string
}

func runUpdate(cmd *cobra.Command, args []string) error {
	currentVersion := version
	if currentVersion == "" || currentVersion == "dev" {
		currentVersion = "dev"
	}

	// Determine target version
	targetVersion := ""
	if len(args) == 1 {
		targetVersion = args[0]
		if !strings.HasPrefix(targetVersion, "v") {
			targetVersion = "v" + targetVersion
		}
	}

	// Fetch latest release (S3 first, GitHub fallback)
	latest, err := fetchLatestRelease()
	if err != nil {
		return fmt.Errorf("failed to check for updates: %w", err)
	}

	latestTag := "v" + latest.Version
	if targetVersion == "" {
		targetVersion = latestTag
	}

	// Display current vs target
	currentDisplay := currentVersion
	if len(currentDisplay) > 0 && currentDisplay[0] != 'v' {
		currentDisplay = "v" + currentDisplay
	}

	// FX3: replace the rc1 single-character != check with the
	// SemVer-aware shouldOfferUpdate helper. autoOffer is the
	// boolean we use when the user has not asked for a specific
	// version (no positional arg) — i.e. the "tene update" / "tene
	// update --check" recommendation path. An explicit positional
	// like `tene update v1.0.13` skips this check (see below) so
	// the user can still roll back deliberately.
	userSuppliedTarget := len(args) == 1
	autoOffer := shouldOfferUpdate(currentDisplay, latestTag, updateFlagIncludePrerelease)

	if flagJSON {
		return printJSON(map[string]any{
			"currentVersion":  currentDisplay,
			"latestVersion":   latestTag,
			"targetVersion":   targetVersion,
			"updateAvailable": autoOffer,
		})
	}

	fmt.Printf("  Current version: %s\n", currentDisplay)
	fmt.Printf("  Latest version:  %s\n", latestTag)

	if currentDisplay == targetVersion {
		fmt.Println("\n  Already up to date.")
		return nil
	}

	// Auto-update path with no version arg: respect the SemVer-aware
	// decision. Display the "you are ahead of the stable channel" or
	// "already up to date" line instead of marching toward a downgrade.
	if !userSuppliedTarget && !autoOffer {
		if semver.IsValid(currentDisplay) && semver.IsValid(latestTag) &&
			semver.Compare(currentDisplay, latestTag) > 0 {
			fmt.Printf("\n  You are on %s, which is newer than the latest stable %s.\n", currentDisplay, latestTag)
			fmt.Println("  Use 'tene update --include-prerelease' to receive prerelease updates,")
			fmt.Println("  or 'tene update vX.Y.Z' to install a specific version.")
		} else {
			fmt.Println("\n  Already up to date.")
		}
		return nil
	}

	fmt.Printf("  Target version:  %s\n", targetVersion)

	if updateFlagCheck {
		if autoOffer {
			fmt.Printf("\n  Update available! Run 'tene update' to install %s.\n", latestTag)
		}
		return nil
	}

	// Confirm update
	if isTerminal() && !flagQuiet {
		fmt.Printf("\n  Update to %s? (y/N) ", targetVersion)
		var answer string
		_, _ = fmt.Scanln(&answer)
		if strings.ToLower(strings.TrimSpace(answer)) != "y" {
			fmt.Println("  Update cancelled.")
			return nil
		}
	}

	// Resolve current binary path
	binPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("cannot determine binary path: %w", err)
	}

	// Download release binary from S3
	versionNum := strings.TrimPrefix(targetVersion, "v")
	assetName := fmt.Sprintf("tene_%s_%s_%s.tar.gz", versionNum, runtime.GOOS, runtime.GOARCH)
	downloadURL := fmt.Sprintf("%s/v%s/%s", releaseBaseURL, versionNum, assetName)

	fmt.Printf("\n  Downloading %s...\n", assetName)

	// Download tar.gz to temp file
	tmpArchive := filepath.Join(os.TempDir(), assetName)
	if err := downloadFile(downloadURL, tmpArchive); err != nil {
		return fmt.Errorf("download failed: %w", err)
	}
	defer func() { _ = os.Remove(tmpArchive) }()

	// Verify checksum of the archive
	checksumsURL := fmt.Sprintf("%s/v%s/checksums.txt", releaseBaseURL, versionNum)
	if err := verifyChecksum(tmpArchive, checksumsURL, assetName); err != nil {
		return fmt.Errorf("checksum verification failed: %w", err)
	}

	// Extract binary from archive
	tmpBin, err := extractBinaryFromTarGz(tmpArchive)
	if err != nil {
		return fmt.Errorf("extraction failed: %w", err)
	}
	defer func() { _ = os.Remove(tmpBin) }()

	// Replace current binary
	if err := replaceBinary(binPath, tmpBin); err != nil {
		return fmt.Errorf("failed to replace binary: %w\n\n  Try manually: curl -sSfL https://tene.sh/install.sh | sh", err)
	}

	fmt.Printf("\n  Updated to %s.\n", targetVersion)
	fmt.Printf("  Run 'tene version' to verify.\n")
	fmt.Printf("\n  Binary location: %s\n", binPath)

	return nil
}

// fetchLatestRelease tries S3 first, falls back to GitHub API.
func fetchLatestRelease() (*releaseInfo, error) {
	info, err := fetchFromS3()
	if err == nil {
		return info, nil
	}

	info, err = fetchFromGitHub()
	if err != nil {
		return nil, fmt.Errorf("cannot check for updates (try: curl -sSfL https://tene.sh/install.sh | sh): %w", err)
	}
	return info, nil
}

func fetchFromS3() (*releaseInfo, error) {
	url := releaseBaseURL + "/LATEST_VERSION"

	resp, err := http.Get(url) //nolint:gosec
	if err != nil {
		return nil, fmt.Errorf("s3 version check failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("s3 returned HTTP %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read version: %w", err)
	}

	ver := strings.TrimSpace(string(body))
	return &releaseInfo{
		Version:     ver,
		DownloadURL: fmt.Sprintf("%s/v%s/", releaseBaseURL, ver),
	}, nil
}

type githubRelease struct {
	TagName string `json:"tag_name"`
	HTMLURL string `json:"html_url"`
}

func fetchFromGitHub() (*releaseInfo, error) {
	url := "https://api.github.com/repos/tene-ai/tene/releases/latest"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", fmt.Sprintf("tene/%s (%s/%s)", version, runtime.GOOS, runtime.GOARCH))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cannot reach GitHub API: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("GitHub API returned %d", resp.StatusCode)
	}

	var release githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	ver := strings.TrimPrefix(release.TagName, "v")
	return &releaseInfo{
		Version: ver,
		HTMLURL: release.HTMLURL,
	}, nil
}

// downloadFile downloads a URL to a local file path.
func downloadFile(url, dest string) error {
	resp, err := http.Get(url) //nolint:gosec
	if err != nil {
		return fmt.Errorf("cannot download: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download returned HTTP %d (check version exists)", resp.StatusCode)
	}

	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer func() { _ = out.Close() }()

	if _, err := io.Copy(out, resp.Body); err != nil {
		return err
	}
	return nil
}

// verifyChecksum downloads checksums.txt and verifies the SHA-256 of the archive file.
func verifyChecksum(archivePath, checksumsURL, assetName string) error {
	resp, err := http.Get(checksumsURL) //nolint:gosec
	if err != nil {
		return fmt.Errorf("cannot download checksums: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		// Checksum file not available — skip verification (backward compat)
		return nil
	}

	var expected string
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, assetName) {
			parts := strings.Fields(line)
			if len(parts) >= 1 {
				expected = parts[0]
				break
			}
		}
	}

	if expected == "" {
		// Asset not found in checksums — skip
		return nil
	}

	f, err := os.Open(archivePath)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return err
	}
	actual := hex.EncodeToString(h.Sum(nil))

	if actual != expected {
		return fmt.Errorf("expected %s, got %s", expected, actual)
	}

	return nil
}

// extractBinaryFromTarGz extracts the tene binary from a local .tar.gz file to a temp file.
func extractBinaryFromTarGz(archivePath string) (string, error) {
	f, err := os.Open(archivePath)
	if err != nil {
		return "", err
	}
	defer func() { _ = f.Close() }()

	gz, err := gzip.NewReader(f)
	if err != nil {
		return "", fmt.Errorf("invalid gzip: %w", err)
	}
	defer func() { _ = gz.Close() }()

	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", fmt.Errorf("invalid tar: %w", err)
		}
		if hdr.Name == "tene" && hdr.Typeflag == tar.TypeReg {
			tmp, err := os.CreateTemp("", "tene-update-*")
			if err != nil {
				return "", err
			}
			if _, err := io.Copy(tmp, tr); err != nil {
				_ = tmp.Close()
				_ = os.Remove(tmp.Name())
				return "", err
			}
			_ = tmp.Close()
			if err := os.Chmod(tmp.Name(), 0755); err != nil {
				_ = os.Remove(tmp.Name())
				return "", err
			}
			return tmp.Name(), nil
		}
	}

	return "", fmt.Errorf("tene binary not found in archive")
}

// replaceBinary atomically replaces the binary at dst with the one at src.
func replaceBinary(dst, src string) error {
	f, err := os.OpenFile(dst, os.O_WRONLY, 0)
	if err != nil {
		return fmt.Errorf("no write permission on %s (try with sudo)", dst)
	}
	_ = f.Close()

	if err := os.Rename(src, dst); err != nil {
		return copyFile(src, dst)
	}
	return nil
}

// copyFile copies src to dst, preserving executable permission.
func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() { _ = in.Close() }()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func() { _ = out.Close() }()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return os.Chmod(dst, 0755)
}
