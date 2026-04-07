package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

	"github.com/spf13/cobra"
	"github.com/tomo-kay/tene/internal/domain"
)

func newLoginCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "login",
		Short: "Sign in to Tene Cloud with GitHub",
		Long:  "Authenticate with Tene Cloud using GitHub OAuth. Required for cloud features (push, pull, sync, team).",
		RunE:  runLogin,
	}
	cmd.Flags().String("api-url", envOrDefault("API_URL", "https://api.tene.sh"), "Tene Cloud API base URL")
	return cmd
}

func runLogin(cmd *cobra.Command, args []string) error {
	apiURL, _ := cmd.Flags().GetString("api-url")

	// Check if already logged in
	token, _ := loadAuthToken()
	if token != "" {
		fmt.Fprintln(cmd.ErrOrStderr(), "  Already logged in. Run 'tene logout' first to switch accounts.")
		return nil
	}

	fmt.Fprintln(cmd.ErrOrStderr(), "  Signing in with GitHub...")

	// Start local callback server
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return fmt.Errorf("login: start callback server: %w", err)
	}
	port := listener.Addr().(*net.TCPAddr).Port

	resultCh := make(chan *loginResult, 1)
	errCh := make(chan error, 1)

	srv := &http.Server{
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/callback" {
				http.NotFound(w, r)
				return
			}

			// Read the tokens from query params (returned by our API after OAuth)
			code := r.URL.Query().Get("code")
			state := r.URL.Query().Get("state")

			if code == "" {
				w.WriteHeader(http.StatusBadRequest)
				fmt.Fprint(w, "Missing code parameter")
				errCh <- fmt.Errorf("missing code in callback")
				return
			}

			// Exchange code via our API
			result, err := exchangeCodeViaAPI(r.Context(), apiURL, code, state)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprintf(w, "Login failed: %v", err)
				errCh <- err
				return
			}

			w.Header().Set("Content-Type", "text/html")
			fmt.Fprint(w, `<html><body><h2>Login successful!</h2><p>You can close this window and return to the terminal.</p></body></html>`)
			resultCh <- result
		}),
	}

	go func() {
		if err := srv.Serve(listener); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	// Open browser to GitHub OAuth (via our API's authorize endpoint)
	authURL := fmt.Sprintf("%s/api/v1/auth/github/authorize?callback_port=%d", apiURL, port)
	fmt.Fprintf(cmd.ErrOrStderr(), "  Opening browser...\n")
	if err := openBrowser(authURL); err != nil {
		fmt.Fprintf(cmd.ErrOrStderr(), "  Could not open browser. Visit this URL:\n  %s\n", authURL)
	}

	fmt.Fprintf(cmd.ErrOrStderr(), "  Waiting for authentication...")

	// Wait for callback (3 min timeout)
	ctx, cancel := context.WithTimeout(cmd.Context(), 3*time.Minute)
	defer cancel()

	select {
	case result := <-resultCh:
		fmt.Fprintf(cmd.ErrOrStderr(), " done\n\n")
		// Save tokens
		if err := saveAuthToken(result.Tokens.AccessToken); err != nil {
			return fmt.Errorf("login: save token: %w", err)
		}
		if err := saveRefreshToken(result.Tokens.RefreshToken); err != nil {
			return fmt.Errorf("login: save refresh token: %w", err)
		}

		fmt.Fprintf(cmd.ErrOrStderr(), "  ✓ Signed in as %s (%s)\n", result.User.Name, result.User.Email)
		fmt.Fprintf(cmd.ErrOrStderr(), "  Plan: %s\n\n", result.User.Plan)
		fmt.Fprintf(cmd.ErrOrStderr(), "  Run 'tene push' to sync your vault.\n")
	case err := <-errCh:
		fmt.Fprintf(cmd.ErrOrStderr(), " failed\n")
		return fmt.Errorf("login: %w", err)
	case <-ctx.Done():
		fmt.Fprintf(cmd.ErrOrStderr(), " timed out\n")
		return fmt.Errorf("login: timed out waiting for authentication")
	}

	// Shutdown callback server
	shutCtx, shutCancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer shutCancel()
	srv.Shutdown(shutCtx)

	return nil
}

type loginResult struct {
	User   domain.User      `json:"user"`
	Tokens domain.TokenPair `json:"tokens"`
}

func exchangeCodeViaAPI(ctx context.Context, apiURL, code, state string) (*loginResult, error) {
	url := fmt.Sprintf("%s/api/v1/auth/github/callback?code=%s&state=%s", apiURL, code, state)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("exchange code: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("exchange code: API returned %d", resp.StatusCode)
	}

	var apiResp struct {
		OK   bool         `json:"ok"`
		Data loginResult  `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("exchange code: decode: %w", err)
	}
	if !apiResp.OK {
		return nil, fmt.Errorf("exchange code: API error")
	}

	return &apiResp.Data, nil
}

func openBrowser(url string) error {
	switch runtime.GOOS {
	case "darwin":
		return exec.Command("open", url).Start()
	case "linux":
		return exec.Command("xdg-open", url).Start()
	case "windows":
		return exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	default:
		return fmt.Errorf("unsupported platform")
	}
}

// Token storage helpers — uses ~/.tene/auth.json
// TODO: Use OS Keychain (go-keyring) for production

func loadAuthToken() (string, error) {
	return loadAuthField("access_token")
}

func saveAuthToken(token string) error {
	return saveAuthField("access_token", token)
}

func saveRefreshToken(token string) error {
	return saveAuthField("refresh_token", token)
}

func loadAuthField(field string) (string, error) {
	data, err := readAuthFile()
	if err != nil {
		return "", err
	}
	return data[field], nil
}

func saveAuthField(field, value string) error {
	data, _ := readAuthFile()
	if data == nil {
		data = make(map[string]string)
	}
	data[field] = value
	return writeAuthFile(data)
}

func clearAuthFile() error {
	return writeAuthFile(map[string]string{})
}

func readAuthFile() (map[string]string, error) {
	path := authFilePath()
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var data map[string]string
	if err := json.Unmarshal(b, &data); err != nil {
		return nil, err
	}
	return data, nil
}

func writeAuthFile(data map[string]string) error {
	path := authFilePath()
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	b, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0600)
}

func authFilePath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".tene", "auth.json")
}
