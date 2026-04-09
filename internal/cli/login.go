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
	"github.com/zalando/go-keyring"

	"github.com/tomo-kay/tene/pkg/domain"
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
		_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "  Already logged in. Run 'tene logout' first to switch accounts.")
		return nil
	}

	_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "  Signing in with GitHub...")

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

			// Read tokens directly from query params (API redirects with tokens after OAuth)
			accessToken := r.URL.Query().Get("access_token")
			refreshToken := r.URL.Query().Get("refresh_token")
			userID := r.URL.Query().Get("user_id")
			plan := r.URL.Query().Get("plan")

			if accessToken == "" {
				w.WriteHeader(http.StatusBadRequest)
				_, _ = fmt.Fprint(w, "Missing access_token parameter")
				errCh <- fmt.Errorf("missing access_token in callback")
				return
			}

			result := &loginResult{
				User:   domain.User{ID: userID, Plan: plan},
				Tokens: domain.TokenPair{AccessToken: accessToken, RefreshToken: refreshToken},
			}

			w.Header().Set("Content-Type", "text/html")
			_, _ = fmt.Fprint(w, `<html><body><h2>Login successful!</h2><p>You can close this window and return to the terminal.</p></body></html>`)
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
	_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "  Opening browser...\n")
	if err := openBrowser(authURL); err != nil {
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "  Could not open browser. Visit this URL:\n  %s\n", authURL)
	}

	_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "  Waiting for authentication...")

	// Wait for callback (3 min timeout)
	ctx, cancel := context.WithTimeout(cmd.Context(), 3*time.Minute)
	defer cancel()

	select {
	case result := <-resultCh:
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), " done\n\n")
		// Save tokens
		if err := saveAuthToken(result.Tokens.AccessToken); err != nil {
			return fmt.Errorf("login: save token: %w", err)
		}
		if err := saveRefreshToken(result.Tokens.RefreshToken); err != nil {
			return fmt.Errorf("login: save refresh token: %w", err)
		}

		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "  ✓ Signed in as %s (%s)\n", result.User.Name, result.User.Email)
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "  Plan: %s\n\n", result.User.Plan)
		if result.User.Plan == "pro" {
			_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "  Dashboard: https://app.tene.sh\n")
			_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "  Run 'tene push' to sync your vault.\n")
		} else {
			_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "  Upgrade to Pro for cloud sync, team sharing, and dashboard.\n")
			_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "  → https://app.tene.sh/upgrade\n")
		}
	case err := <-errCh:
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), " failed\n")
		return fmt.Errorf("login: %w", err)
	case <-ctx.Done():
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), " timed out\n")
		return fmt.Errorf("login: timed out waiting for authentication")
	}

	// Shutdown callback server
	shutCtx, shutCancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer shutCancel()
	_ = srv.Shutdown(shutCtx)

	return nil
}

type loginResult struct {
	User   domain.User      `json:"user"`
	Tokens domain.TokenPair `json:"tokens"`
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

// Token storage helpers
// Uses OS Keychain (go-keyring) by default, falls back to ~/.tene/auth.json
// when --no-keychain is set or keychain is unavailable.

const (
	cloudServiceName = "tene-cloud"
	keyAccessToken   = "access_token"
	keyRefreshToken  = "refresh_token"
)

func loadAuthToken() (string, error) {
	return loadAuthField(keyAccessToken)
}

func saveAuthToken(token string) error {
	return saveAuthField(keyAccessToken, token)
}

func saveRefreshToken(token string) error {
	return saveAuthField(keyRefreshToken, token)
}

func loadAuthField(field string) (string, error) {
	if !flagNoKeychain {
		// Try keychain first
		val, err := keyring.Get(cloudServiceName, field)
		if err == nil {
			return val, nil
		}
		// Keychain error (not ErrNotFound) — fall through to file fallback

		// Try file fallback and migrate to keychain if found
		val, err = loadAuthFieldFromFile(field)
		if err == nil && val != "" {
			// Migrate to keychain
			_ = keyring.Set(cloudServiceName, field, val)
			// Also migrate the other field if present
			otherField := keyRefreshToken
			if field == keyRefreshToken {
				otherField = keyAccessToken
			}
			if otherVal, err := loadAuthFieldFromFile(otherField); err == nil && otherVal != "" {
				_ = keyring.Set(cloudServiceName, otherField, otherVal)
			}
			// Remove auth file after successful migration
			_ = os.Remove(authFilePath())
			return val, nil
		}
		return "", err
	}

	// --no-keychain: use file only
	return loadAuthFieldFromFile(field)
}

func saveAuthField(field, value string) error {
	if !flagNoKeychain {
		if err := keyring.Set(cloudServiceName, field, value); err == nil {
			return nil
		}
		// Keychain failed — fall through to file
	}
	return saveAuthFieldToFile(field, value)
}

func clearAuthTokens() error {
	// Clear from keychain (ignore errors)
	if !flagNoKeychain {
		_ = keyring.Delete(cloudServiceName, keyAccessToken)
		_ = keyring.Delete(cloudServiceName, keyRefreshToken)
	}
	// Also clear file if it exists
	return clearAuthFile()
}

// --- File-based storage (fallback / legacy) ---

func loadAuthFieldFromFile(field string) (string, error) {
	data, err := readAuthFile()
	if err != nil {
		return "", err
	}
	return data[field], nil
}

func saveAuthFieldToFile(field, value string) error {
	data, _ := readAuthFile()
	if data == nil {
		data = make(map[string]string)
	}
	data[field] = value
	return writeAuthFile(data)
}

func clearAuthFile() error {
	path := authFilePath()
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil
	}
	return os.Remove(path)
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
