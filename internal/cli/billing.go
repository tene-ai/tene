package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/spf13/cobra"
)

func newBillingCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "billing",
		Short: "Manage your Tene Cloud subscription",
		Long:  "View subscription status and manage your Pro plan.",
		RunE:  runBillingStatus,
	}

	cmd.AddCommand(newBillingUpgradeCmd())
	cmd.AddCommand(newBillingPortalCmd())
	cmd.Flags().String("api-url", envOrDefault("API_URL", "https://api.tene.sh"), "Tene Cloud API base URL")
	return cmd
}

func newBillingUpgradeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "upgrade",
		Short: "Upgrade to Pro plan ($5/month)",
		RunE:  runBillingUpgrade,
	}
	cmd.Flags().String("api-url", envOrDefault("API_URL", "https://api.tene.sh"), "Tene Cloud API base URL")
	return cmd
}

func newBillingPortalCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "portal",
		Short: "Open subscription management portal",
		RunE:  runBillingPortal,
	}
	cmd.Flags().String("api-url", envOrDefault("API_URL", "https://api.tene.sh"), "Tene Cloud API base URL")
	return cmd
}

func runBillingStatus(cmd *cobra.Command, args []string) error {
	apiURL, _ := cmd.Flags().GetString("api-url")
	token, err := loadAuthToken()
	if err != nil || token == "" {
		return fmt.Errorf("not logged in. Run 'tene login' first")
	}

	result, err := billingAPIGet(apiURL+"/api/v1/billing/subscription", token)
	if err != nil {
		return fmt.Errorf("billing: %w", err)
	}

	if flagJSON {
		return printJSON(result)
	}

	plan, _ := result["plan"].(string)
	status, _ := result["status"].(string)

	fmt.Fprintf(cmd.ErrOrStderr(), "  Plan: %s\n", plan)
	fmt.Fprintf(cmd.ErrOrStderr(), "  Status: %s\n", status)

	if plan == "free" {
		fmt.Fprintln(cmd.ErrOrStderr(), "\n  Upgrade to Pro ($5/month) for cloud sync, backup, and team features.")
		fmt.Fprintln(cmd.ErrOrStderr(), "  Run: tene billing upgrade")
	}

	return nil
}

func runBillingUpgrade(cmd *cobra.Command, args []string) error {
	apiURL, _ := cmd.Flags().GetString("api-url")
	token, err := loadAuthToken()
	if err != nil || token == "" {
		return fmt.Errorf("not logged in. Run 'tene login' first")
	}

	// Get user email from /auth/me
	me, err := billingAPIGet(apiURL+"/api/v1/auth/me", token)
	if err != nil {
		return fmt.Errorf("billing: get user: %w", err)
	}

	email, _ := me["email"].(string)
	if email == "" {
		email, _ = me["user_id"].(string) // fallback
	}

	// Create checkout session
	body, _ := json.Marshal(map[string]string{"email": email})
	result, err := billingAPIPost(apiURL+"/api/v1/billing/checkout", token, body)
	if err != nil {
		return fmt.Errorf("billing: %w", err)
	}

	checkoutURL, _ := result["checkout_url"].(string)
	if checkoutURL == "" {
		return fmt.Errorf("billing: no checkout URL returned")
	}

	fmt.Fprintf(cmd.ErrOrStderr(), "  Opening checkout...\n")
	if err := openBrowser(checkoutURL); err != nil {
		fmt.Fprintf(cmd.ErrOrStderr(), "  Could not open browser. Visit:\n  %s\n", checkoutURL)
	} else {
		fmt.Fprintf(cmd.ErrOrStderr(), "  ✓ Checkout opened in browser\n")
	}

	return nil
}

func runBillingPortal(cmd *cobra.Command, args []string) error {
	apiURL, _ := cmd.Flags().GetString("api-url")
	token, err := loadAuthToken()
	if err != nil || token == "" {
		return fmt.Errorf("not logged in. Run 'tene login' first")
	}

	result, err := billingAPIPost(apiURL+"/api/v1/billing/portal", token, nil)
	if err != nil {
		return fmt.Errorf("billing: %w", err)
	}

	portalURL, _ := result["portal_url"].(string)
	if portalURL == "" {
		return fmt.Errorf("billing: no portal URL. Are you on Pro plan?")
	}

	fmt.Fprintf(cmd.ErrOrStderr(), "  Opening subscription portal...\n")
	if err := openBrowser(portalURL); err != nil {
		fmt.Fprintf(cmd.ErrOrStderr(), "  Visit:\n  %s\n", portalURL)
	}

	return nil
}

func billingAPIGet(url, token string) (map[string]any, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	return doAPIRequest(req)
}

func billingAPIPost(url, token string, body []byte) (map[string]any, error) {
	var req *http.Request
	var err error
	if body != nil {
		req, err = http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	} else {
		req, err = http.NewRequest(http.MethodPost, url, nil)
	}
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	return doAPIRequest(req)
}

var cliHTTPClient = &http.Client{Timeout: 30 * time.Second}

func doAPIRequest(req *http.Request) (map[string]any, error) {
	resp, err := cliHTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var apiResp struct {
		OK   bool           `json:"ok"`
		Data map[string]any `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	if !apiResp.OK {
		return nil, fmt.Errorf("API error (status %d)", resp.StatusCode)
	}
	return apiResp.Data, nil
}

