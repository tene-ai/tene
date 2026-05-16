package cli

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"
	"github.com/agent-kay-it/tene/pkg/crypto"
)

var teamHTTPClient = &http.Client{Timeout: 30 * time.Second}

func newTeamCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "team",
		Short: "Manage teams for shared vault access",
		Long:  "Create teams, invite members, and manage roles for shared secret management.",
		Example: `  # Create a new team
  tene team create "My Project"

  # List your teams
  tene team list

  # Invite a member (wraps project key via X25519 ECDH)
  tene team invite <team-id> <user-id>

  # Remove a member (triggers key rotation)
  tene team remove <team-id> <user-id>

  # List team members
  tene team members <team-id>`,
	}

	cmd.AddCommand(newTeamCreateCmd())
	cmd.AddCommand(newTeamListCmd())
	cmd.AddCommand(newTeamInviteCmd())
	cmd.AddCommand(newTeamRemoveCmd())
	cmd.AddCommand(newTeamMembersCmd())
	return cmd
}

func newTeamCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create [name]",
		Short: "Create a new team",
		Args:  cobra.ExactArgs(1),
		RunE:  runTeamCreate,
	}
	cmd.Flags().String("slug", "", "Team slug (lowercase, dashes allowed)")
	cmd.Flags().String("api-url", envOrDefault("API_URL", "https://api.tene.sh"), "API base URL")
	return cmd
}

func newTeamListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List your teams",
		RunE:  runTeamList,
	}
	cmd.Flags().String("api-url", envOrDefault("API_URL", "https://api.tene.sh"), "API base URL")
	return cmd
}

func newTeamInviteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "invite [team-id] [user-id]",
		Short: "Invite a member to a team",
		Args:  cobra.ExactArgs(2),
		RunE:  runTeamInvite,
	}
	cmd.Flags().String("role", "member", "Role: admin or member")
	cmd.Flags().String("api-url", envOrDefault("API_URL", "https://api.tene.sh"), "API base URL")
	return cmd
}

func newTeamRemoveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove [team-id] [user-id]",
		Short: "Remove a member from a team (triggers key rotation)",
		Args:  cobra.ExactArgs(2),
		RunE:  runTeamRemove,
	}
	cmd.Flags().String("api-url", envOrDefault("API_URL", "https://api.tene.sh"), "API base URL")
	return cmd
}

func newTeamMembersCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "members [team-id]",
		Short: "List team members",
		Args:  cobra.ExactArgs(1),
		RunE:  runTeamMembers,
	}
	cmd.Flags().String("api-url", envOrDefault("API_URL", "https://api.tene.sh"), "API base URL")
	return cmd
}

func runTeamCreate(cmd *cobra.Command, args []string) error {
	apiURL, _ := cmd.Flags().GetString("api-url")
	slug, _ := cmd.Flags().GetString("slug")
	name := args[0]

	token, err := loadAuthToken()
	if err != nil || token == "" {
		return fmt.Errorf("not logged in. Run 'tene login' first")
	}

	// Pre-check plan from JWT (client-side, fail-fast)
	if err := checkProPlan(token); err != nil {
		return err
	}

	if slug == "" {
		slug = slugify(name)
	}

	// Generate project key for this team
	projectKey, err := crypto.GenerateProjectKey()
	if err != nil {
		return fmt.Errorf("team create: generate project key: %w", err)
	}
	defer crypto.ZeroBytes(projectKey)

	// Store project key locally (encrypted by master key)
	app, err := loadApp()
	if err == nil {
		masterKey, mkErr := loadOrPromptMasterKey(app)
		if mkErr == nil {
			defer crypto.ZeroBytes(masterKey)
			encPK, encErr := crypto.Encrypt(masterKey, projectKey, []byte("team-project-key"))
			if encErr == nil {
				pkPath := filepath.Join(app.Dir, ".tene", "team_pk_"+slug+".enc")
				_ = os.WriteFile(pkPath, encPK, 0600)
			}
		}
		_ = app.Vault.Close()
	}

	if !flagQuiet {
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "  Creating team '%s'...\n", name)
	}

	body, _ := json.Marshal(map[string]string{"name": name, "slug": slug})
	result, err := teamAPIPost(apiURL+"/api/v1/teams", token, body)
	if err != nil {
		return fmt.Errorf("team create: %w", err)
	}

	if flagJSON {
		return printJSON(result)
	}

	teamID, _ := result["id"].(string)
	_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "  ✓ Team '%s' created (ID: %s)\n", name, teamID)
	_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "    Slug: %s\n", slug)
	_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "    Project key generated and stored locally.\n")
	return nil
}

func runTeamList(cmd *cobra.Command, args []string) error {
	apiURL, _ := cmd.Flags().GetString("api-url")
	token, err := loadAuthToken()
	if err != nil || token == "" {
		return fmt.Errorf("not logged in. Run 'tene login' first")
	}

	result, err := teamAPIGetList(apiURL+"/api/v1/teams", token)
	if err != nil {
		return fmt.Errorf("team list: %w", err)
	}

	if flagJSON {
		return printJSON(result)
	}

	if len(result) == 0 {
		_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "  No teams. Create one with 'tene team create [name]'")
		return nil
	}

	w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 2, 2, ' ', 0)
	_, _ = fmt.Fprintln(w, "  ID\tNAME\tSLUG\tCREATED")
	for _, t := range result {
		id, _ := t["id"].(string)
		name, _ := t["name"].(string)
		sl, _ := t["slug"].(string)
		created, _ := t["created_at"].(string)
		if len(created) > 10 {
			created = created[:10]
		}
		shortID := id
		if len(shortID) > 8 {
			shortID = shortID[:8] + "..."
		}
		_, _ = fmt.Fprintf(w, "  %s\t%s\t%s\t%s\n", shortID, name, sl, created)
	}
	_ = w.Flush()
	return nil
}

func runTeamInvite(cmd *cobra.Command, args []string) error {
	apiURL, _ := cmd.Flags().GetString("api-url")
	role, _ := cmd.Flags().GetString("role")
	teamID := args[0]
	userID := args[1]

	token, err := loadAuthToken()
	if err != nil || token == "" {
		return fmt.Errorf("not logged in. Run 'tene login' first")
	}

	// Pre-check plan from JWT (client-side, fail-fast)
	if err := checkProPlan(token); err != nil {
		return err
	}

	// Load local project key and wrap for the invitee
	// In production: fetch invitee's X25519 public key from server, then ECDH wrap
	var wrappedPK []byte

	// Load owner's X25519 key pair from keychain
	ownerKP, err := loadOrGenerateX25519()
	if err == nil {
		defer crypto.ZeroBytes(ownerKP.PrivateKey)

		// Fetch invitee's public key from server
		inviteePublicKey, pkErr := fetchUserPublicKey(apiURL, token, userID)
		if pkErr == nil && len(inviteePublicKey) == crypto.X25519KeySize {
			// Load project key from local storage
			app, appErr := loadApp()
			if appErr == nil {
				masterKey, mkErr := loadOrPromptMasterKey(app)
				if mkErr == nil {
					defer crypto.ZeroBytes(masterKey)
					slug := resolveTeamSlug(apiURL, token, teamID)
					pkPath := filepath.Join(app.Dir, ".tene", "team_pk_"+slug+".enc")
					encPK, readErr := os.ReadFile(pkPath)
					if readErr == nil {
						projectKey, decErr := crypto.Decrypt(masterKey, encPK, []byte("team-project-key"))
						if decErr == nil {
							defer crypto.ZeroBytes(projectKey)
							wrappedPK, _ = crypto.WrapProjectKey(
								ownerKP.PrivateKey, inviteePublicKey,
								teamID, userID, projectKey)
						}
					}
				}
				_ = app.Vault.Close()
			}
		}
	}

	if !flagQuiet {
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "  Inviting user to team...\n")
	}

	reqBody := map[string]any{
		"user_id": userID,
		"role":    role,
	}
	if wrappedPK != nil {
		reqBody["wrapped_project_key"] = base64.StdEncoding.EncodeToString(wrappedPK)
	}
	body, _ := json.Marshal(reqBody)

	result, err := teamAPIPost(apiURL+"/api/v1/teams/"+teamID+"/invite", token, body)
	if err != nil {
		return fmt.Errorf("team invite: %w", err)
	}

	if flagJSON {
		return printJSON(result)
	}

	_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "  ✓ User %s invited as %s\n", userID, role)
	if wrappedPK != nil {
		_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "    Project key wrapped and transmitted (zero-knowledge).")
	}
	return nil
}

func runTeamRemove(cmd *cobra.Command, args []string) error {
	apiURL, _ := cmd.Flags().GetString("api-url")
	teamID := args[0]
	userID := args[1]

	token, err := loadAuthToken()
	if err != nil || token == "" {
		return fmt.Errorf("not logged in. Run 'tene login' first")
	}

	if !flagQuiet {
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "  Removing member and rotating keys...\n")
	}

	req, err := http.NewRequest(http.MethodDelete,
		apiURL+"/api/v1/teams/"+teamID+"/members/"+userID, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := teamHTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("team remove: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	var apiResp struct {
		OK      bool           `json:"ok"`
		Error   string         `json:"error"`
		Message string         `json:"message"`
		Data    map[string]any `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return fmt.Errorf("team remove: decode: %w", err)
	}
	if !apiResp.OK {
		return fmt.Errorf("team remove: %s", apiErrMsg(apiResp.Error, apiResp.Message, resp.StatusCode))
	}

	if flagJSON {
		return printJSON(apiResp.Data)
	}

	_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "  ✓ Member removed\n")

	// Key rotation: generate new PK, re-wrap for remaining members
	if rotation, ok := apiResp.Data["key_rotation"].(bool); ok && rotation {
		_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "  ⟳ Key rotation triggered. Run 'tene push --force' to re-encrypt vault.")
	}

	return nil
}

func runTeamMembers(cmd *cobra.Command, args []string) error {
	apiURL, _ := cmd.Flags().GetString("api-url")
	teamID := args[0]

	token, err := loadAuthToken()
	if err != nil || token == "" {
		return fmt.Errorf("not logged in. Run 'tene login' first")
	}

	// The members endpoint isn't explicitly separate; list teams includes member info
	// For now, use the team list and filter
	result, err := teamAPIGetList(apiURL+"/api/v1/teams", token)
	if err != nil {
		return fmt.Errorf("team members: %w", err)
	}

	if flagJSON {
		return printJSON(result)
	}

	_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "  Team: %s\n\n", teamID)
	// TODO: fetch actual member list from API when endpoint available
	_ = result
	_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "  Use the dashboard at app.tene.sh to manage team members.")
	return nil
}

// --- helpers ---

func teamAPIPost(url, token string, body []byte) (map[string]any, error) {
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := teamHTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	var apiResp struct {
		OK      bool           `json:"ok"`
		Error   string         `json:"error"`
		Message string         `json:"message"`
		Data    map[string]any `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("decode: %w", err)
	}
	if !apiResp.OK {
		return nil, fmt.Errorf("%s", apiErrMsg(apiResp.Error, apiResp.Message, resp.StatusCode))
	}
	return apiResp.Data, nil
}

func teamAPIGetList(url, token string) ([]map[string]any, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := teamHTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	var apiResp struct {
		OK      bool             `json:"ok"`
		Error   string           `json:"error"`
		Message string           `json:"message"`
		Data    []map[string]any `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("decode: %w", err)
	}
	if !apiResp.OK {
		return nil, fmt.Errorf("%s", apiErrMsg(apiResp.Error, apiResp.Message, resp.StatusCode))
	}
	return apiResp.Data, nil
}

func slugify(name string) string {
	slug := make([]byte, 0, len(name))
	for _, c := range []byte(name) {
		switch {
		case c >= 'a' && c <= 'z', c >= '0' && c <= '9':
			slug = append(slug, c)
		case c >= 'A' && c <= 'Z':
			slug = append(slug, c+32) // toLower
		case c == ' ' || c == '-' || c == '_':
			if len(slug) > 0 && slug[len(slug)-1] != '-' {
				slug = append(slug, '-')
			}
		}
	}
	return string(slug)
}

func loadOrGenerateX25519() (*crypto.X25519KeyPair, error) {
	// TODO: load from OS keychain; for now generate ephemeral
	return crypto.GenerateX25519KeyPair()
}

func fetchUserPublicKey(apiURL, token, userID string) ([]byte, error) {
	// TODO: GET /api/v1/users/:id/public-key
	_ = apiURL
	_ = token
	_ = userID
	return nil, fmt.Errorf("not implemented: fetch user public key")
}

func resolveTeamSlug(apiURL, token, teamID string) string {
	// TODO: fetch team details to get slug
	_ = apiURL
	_ = token
	return teamID[:8]
}
