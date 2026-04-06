package cli

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
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
			"currentVersion": currentDisplay,
			"latestVersion":  latest.TagName,
			"targetVersion":  targetVersion,
			"updateAvailable": currentDisplay != latest.TagName && currentDisplay != "vdev",
			"releaseUrl":     latest.HTMLURL,
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

	// Check if go is available
	goPath, err := exec.LookPath("go")
	if err != nil {
		// No Go installed — suggest manual download
		fmt.Println()
		fmt.Printf("  Go is not installed. Download the binary manually:\n")
		fmt.Printf("  https://github.com/tomo-kay/tene/releases/tag/%s\n", targetVersion)
		fmt.Println()
		fmt.Printf("  Or install Go first: https://go.dev/dl/\n")
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

	// Run go install
	installPath := fmt.Sprintf("github.com/tomo-kay/tene/cmd/tene@%s", targetVersion)
	fmt.Printf("\n  Installing %s...\n", installPath)

	goInstall := exec.Command(goPath, "install", installPath)
	goInstall.Stdout = os.Stdout
	goInstall.Stderr = os.Stderr
	goInstall.Env = os.Environ()

	if err := goInstall.Run(); err != nil {
		return fmt.Errorf("update failed: %w\n\n  Try manually: go install %s", err, installPath)
	}

	fmt.Printf("\n  Updated to %s.\n", targetVersion)
	fmt.Printf("  Run 'tene version' to verify.\n")

	// Show PATH hint if needed
	gobin := os.Getenv("GOBIN")
	if gobin == "" {
		gopath := os.Getenv("GOPATH")
		if gopath == "" {
			home, _ := os.UserHomeDir()
			gopath = home + "/go"
		}
		gobin = gopath + "/bin"
	}
	fmt.Printf("\n  Binary location: %s/tene\n", gobin)

	return nil
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
