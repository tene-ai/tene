package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
)

// PKCEParams holds a PKCE code_verifier and its S256 code_challenge.
type PKCEParams struct {
	CodeVerifier  string
	CodeChallenge string
}

// GeneratePKCE creates a new PKCE verifier and its S256 challenge.
func GeneratePKCE() (*PKCEParams, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return nil, fmt.Errorf("auth: generate pkce verifier: %w", err)
	}
	verifier := base64.RawURLEncoding.EncodeToString(b)
	sum := sha256.Sum256([]byte(verifier))
	challenge := base64.RawURLEncoding.EncodeToString(sum[:])
	return &PKCEParams{
		CodeVerifier:  verifier,
		CodeChallenge: challenge,
	}, nil
}

// GitHubUser represents user info from GitHub's API.
type GitHubUser struct {
	ID        int64  `json:"id"`
	Login     string `json:"login"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	AvatarURL string `json:"avatar_url"`
}

// OAuthService manages OAuth flows.
type OAuthService struct {
	githubConfig *oauth2.Config
}

// NewOAuthService creates an OAuth service with GitHub configuration.
func NewOAuthService(githubClientID, githubClientSecret, callbackBase string) *OAuthService {
	return &OAuthService{
		githubConfig: &oauth2.Config{
			ClientID:     githubClientID,
			ClientSecret: githubClientSecret,
			Scopes:       []string{"user:email"},
			Endpoint:     github.Endpoint,
			RedirectURL:  callbackBase + "/api/v1/auth/github/callback",
		},
	}
}

// GitHubAuthURL generates the GitHub OAuth authorization URL with a random state and PKCE challenge.
// It returns the authorization URL, the state value, and the PKCE params.
func (s *OAuthService) GitHubAuthURL() (url, state string, pkce *PKCEParams, err error) {
	state, err = generateState()
	if err != nil {
		return "", "", nil, err
	}
	pkce, err = GeneratePKCE()
	if err != nil {
		return "", "", nil, err
	}
	url = s.githubConfig.AuthCodeURL(state, oauth2.AccessTypeOnline,
		oauth2.SetAuthURLParam("code_challenge", pkce.CodeChallenge),
		oauth2.SetAuthURLParam("code_challenge_method", "S256"),
	)
	return url, state, pkce, nil
}

// ExchangeGitHubCode exchanges an authorization code for a GitHub user.
// Pass the PKCE code_verifier that was used to generate the authorization URL.
func (s *OAuthService) ExchangeGitHubCode(ctx context.Context, code, codeVerifier string) (*GitHubUser, error) {
	opts := []oauth2.AuthCodeOption{}
	if codeVerifier != "" {
		opts = append(opts, oauth2.SetAuthURLParam("code_verifier", codeVerifier))
	}
	token, err := s.githubConfig.Exchange(ctx, code, opts...)
	if err != nil {
		return nil, fmt.Errorf("auth: github code exchange: %w", err)
	}

	client := s.githubConfig.Client(ctx, token)

	// Fetch user info
	resp, err := client.Get("https://api.github.com/user")
	if err != nil {
		return nil, fmt.Errorf("auth: github user fetch: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024)) // M-06: limit error body read
		return nil, fmt.Errorf("auth: github user API %d: %s", resp.StatusCode, body)
	}

	// M-06: Limit response body to 1MB to prevent memory exhaustion
	var user GitHubUser
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&user); err != nil {
		return nil, fmt.Errorf("auth: github user decode: %w", err)
	}

	// Fetch primary email if not public
	if user.Email == "" {
		email, err := s.fetchGitHubPrimaryEmail(ctx, client)
		if err == nil {
			user.Email = email
		}
	}

	return &user, nil
}

func (s *OAuthService) fetchGitHubPrimaryEmail(ctx context.Context, client *http.Client) (string, error) {
	resp, err := client.Get("https://api.github.com/user/emails")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var emails []struct {
		Email   string `json:"email"`
		Primary bool   `json:"primary"`
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&emails); err != nil {
		return "", err
	}

	for _, e := range emails {
		if e.Primary {
			return e.Email, nil
		}
	}
	if len(emails) > 0 {
		return emails[0].Email, nil
	}
	return "", fmt.Errorf("no email found")
}

func generateState() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("auth: generate state: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
