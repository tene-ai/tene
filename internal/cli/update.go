package cli

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
)

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

var updateFlagCheck bool

func init() {
	updateCmd.Flags().BoolVar(&updateFlagCheck, "check", false, "Check for updates without installing")
}

type githubRelease struct {
	TagName string `json:"tag_name"`
	HTMLURL string `json:"html_url"`
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

	// Fetch latest release from GitHub API
	latest, err := fetchLatestRelease()
	if err != nil {
		return fmt.Errorf("failed to check for updates: %w", err)
	}

	if targetVersion == "" {
		targetVersion = latest.TagName
	}

	// Display current vs target
	currentDisplay := currentVersion
	if len(currentDisplay) > 0 && currentDisplay[0] != 'v' {
		currentDisplay = "v" + currentDisplay
	}

	if flagJSON {
		return printJSON(map[string]any{
			"currentVersion":  currentDisplay,
			"latestVersion":   latest.TagName,
			"targetVersion":   targetVersion,
			"updateAvailable": currentDisplay != latest.TagName && currentDisplay != "vdev",
			"releaseUrl":      latest.HTMLURL,
		})
	}

	fmt.Printf("  Current version: %s\n", currentDisplay)
	fmt.Printf("  Latest version:  %s\n", latest.TagName)

	if currentDisplay == targetVersion {
		fmt.Println("\n  Already up to date.")
		return nil
	}

	fmt.Printf("  Target version:  %s\n", targetVersion)

	if updateFlagCheck {
		if currentDisplay != latest.TagName && currentDisplay != "vdev" {
			fmt.Printf("\n  Update available! Run 'tene update' to install %s.\n", latest.TagName)
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

	// Download release binary from GitHub
	versionNum := strings.TrimPrefix(targetVersion, "v")
	assetName := fmt.Sprintf("tene_%s_%s_%s.tar.gz", versionNum, runtime.GOOS, runtime.GOARCH)
	downloadURL := fmt.Sprintf("https://github.com/tomo-kay/tene/releases/download/%s/%s", targetVersion, assetName)

	fmt.Printf("\n  Downloading %s...\n", assetName)

	tmpBin, err := downloadBinaryFromTarGz(downloadURL)
	if err != nil {
		return fmt.Errorf("download failed: %w", err)
	}
	defer os.Remove(tmpBin)

	// Replace current binary
	if err := replaceBinary(binPath, tmpBin); err != nil {
		return fmt.Errorf("failed to replace binary: %w\n\n  Try manually: curl -sSfL https://tene.sh/install.sh | sh", err)
	}

	fmt.Printf("\n  Updated to %s.\n", targetVersion)
	fmt.Printf("  Run 'tene version' to verify.\n")
	fmt.Printf("\n  Binary location: %s\n", binPath)

	return nil
}

// downloadBinaryFromTarGz downloads a .tar.gz release asset and extracts the tene binary to a temp file.
func downloadBinaryFromTarGz(url string) (string, error) {
	resp, err := http.Get(url) //nolint:gosec
	if err != nil {
		return "", fmt.Errorf("cannot download: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("download returned HTTP %d (check version exists)", resp.StatusCode)
	}

	gz, err := gzip.NewReader(resp.Body)
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
	// Check write permission by opening for write
	f, err := os.OpenFile(dst, os.O_WRONLY, 0)
	if err != nil {
		return fmt.Errorf("no write permission on %s (try with sudo)", dst)
	}
	_ = f.Close()

	// Rename is atomic on same filesystem; fall back to copy if cross-device
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

func fetchLatestRelease() (*githubRelease, error) {
	url := "https://api.github.com/repos/tomo-kay/tene/releases/latest"

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

	return &release, nil
}
