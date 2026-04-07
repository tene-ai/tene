package cli

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/spf13/cobra"
)

func newLogoutCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "logout",
		Short: "Sign out from Tene Cloud",
		Long:  "Revoke tokens and remove local credentials. Your local vault data is preserved.",
		RunE:  runLogout,
	}
	cmd.Flags().String("api-url", envOrDefault("API_URL", "https://api.tene.sh"), "Tene Cloud API base URL")
	return cmd
}

func runLogout(cmd *cobra.Command, args []string) error {
	apiURL, _ := cmd.Flags().GetString("api-url")

	token, _ := loadAuthToken()
	if token == "" {
		_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "  Not logged in.")
		return nil
	}

	// Attempt to revoke the token on the server; failure is non-fatal.
	if err := callSignout(cmd.Context(), apiURL, token); err != nil {
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "  Warning: server signout failed (%v). Clearing local credentials anyway.\n", err)
	}

	if err := clearAuthFile(); err != nil {
		return fmt.Errorf("logout: clear credentials: %w", err)
	}

	_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "  Signed out. Your local vault data is preserved.")
	return nil
}

// callSignout calls POST /api/v1/auth/signout with the Bearer token.
func callSignout(ctx context.Context, apiURL, token string) error {
	reqCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	url := apiURL + "/api/v1/auth/signout"
	req, err := http.NewRequestWithContext(reqCtx, http.MethodPost, url, nil)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned %d", resp.StatusCode)
	}
	return nil
}
