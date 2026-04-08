package handler

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/tomo-kay/tene/internal/api/middleware"
	"github.com/tomo-kay/tene/internal/api/response"
	"github.com/tomo-kay/tene/internal/auth"
	"github.com/tomo-kay/tene/internal/domain"
)

const (
	stateTTL   = 5 * time.Minute
	authCodeTTL = 30 * time.Second
)

type stateEntry struct {
	codeVerifier string
	redirectTo   string // "dashboard", "cli:{port}", or empty (API-only)
	expiresAt    time.Time
}

// authCodeEntry stores a one-time auth code for dashboard token exchange.
type authCodeEntry struct {
	accessToken  string
	refreshToken string
	expiresAt    time.Time
	used         bool
}

// refreshEntry stores metadata for a refresh token.
type refreshEntry struct {
	userID    string
	plan      string
	tokenHash string
	family    string    // H-04: token family ID for reuse detection
	expiresAt time.Time
}

// AuthUserStore defines user operations for the auth handler.
// Optional: when nil, /auth/me returns JWT claims only and users are not persisted.
type AuthUserStore interface {
	GetUserByID(ctx context.Context, id string) (*domain.User, error)
	UpsertUser(ctx context.Context, u *domain.User) error
}

// AuthHandler handles OAuth and token endpoints.
type AuthHandler struct {
	oauth        *auth.OAuthService
	jwt          *auth.JWTService
	dashboardURL string // For post-login redirect (e.g. "https://app.tene.sh")
	userStore    AuthUserStore // optional, nil for in-memory mode
	mu           sync.RWMutex
	states       map[string]stateEntry     // in-memory state store with TTL (replace with Redis in prod)
	refresh      map[string]refreshEntry   // in-memory refresh token store (replace with DB in prod)
	authCodes    map[string]authCodeEntry   // one-time auth codes for dashboard token exchange
}

// NewAuthHandler creates an auth handler.
func NewAuthHandler(oauth *auth.OAuthService, jwt *auth.JWTService, dashboardURL string) *AuthHandler {
	h := &AuthHandler{
		oauth:        oauth,
		jwt:          jwt,
		dashboardURL: dashboardURL,
		states:       make(map[string]stateEntry),
		refresh:      make(map[string]refreshEntry),
		authCodes:    make(map[string]authCodeEntry),
	}
	go h.cleanupLoop()
	return h
}

// SetUserStore sets an optional user store for enriching /auth/me responses.
func (h *AuthHandler) SetUserStore(store AuthUserStore) {
	h.userStore = store
}

// cleanupLoop removes expired state and refresh entries every minute.
func (h *AuthHandler) cleanupLoop() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		now := time.Now()
		h.mu.Lock()
		for k, v := range h.states {
			if now.After(v.expiresAt) {
				delete(h.states, k)
			}
		}
		for k, v := range h.refresh {
			if now.After(v.expiresAt) {
				delete(h.refresh, k)
			}
		}
		for k, v := range h.authCodes {
			if now.After(v.expiresAt) {
				delete(h.authCodes, k)
			}
		}
		h.mu.Unlock()
	}
}

// hashToken returns the SHA-256 hex digest of a token string.
func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

// GitHubAuthorize redirects to GitHub OAuth page.
func (h *AuthHandler) GitHubAuthorize(c echo.Context) error {
	url, state, pkce, err := h.oauth.GitHubAuthURL()
	if err != nil {
		return response.Err(c, err)
	}
	// Determine redirect target
	redirectTo := c.QueryParam("redirect") // "dashboard" or empty
	if port := c.QueryParam("callback_port"); port != "" {
		redirectTo = "cli:" + port // CLI local callback server
	}

	h.mu.Lock()
	h.states[state] = stateEntry{
		codeVerifier: pkce.CodeVerifier,
		redirectTo:   redirectTo,
		expiresAt:    time.Now().Add(stateTTL),
	}
	h.mu.Unlock()
	return c.Redirect(http.StatusTemporaryRedirect, url)
}

// GitHubCallback handles the OAuth callback from GitHub.
func (h *AuthHandler) GitHubCallback(c echo.Context) error {
	code := c.QueryParam("code")
	state := c.QueryParam("state")

	if code == "" {
		return response.Err(c, domain.ErrInvalidOAuthCode)
	}

	// Validate state (CSRF protection)
	h.mu.Lock()
	entry, ok := h.states[state]
	if ok {
		delete(h.states, state)
	}
	h.mu.Unlock()

	if !ok || time.Now().After(entry.expiresAt) {
		slog.Warn("auth.login.failed", "reason", "invalid_state", "ip", c.RealIP())
		return response.ErrMsg(c, http.StatusBadRequest, "INVALID_STATE", "invalid oauth state")
	}

	// Exchange code for GitHub user info (pass PKCE verifier)
	ghUser, err := h.oauth.ExchangeGitHubCode(c.Request().Context(), code, entry.codeVerifier)
	if err != nil {
		return response.Err(c, domain.ErrInvalidOAuthCode)
	}

	// Create user object from GitHub data
	user := &domain.User{
		ID:           generateUserID(ghUser.ID),
		Email:        ghUser.Email,
		Name:         ghUser.Name,
		AuthProvider: "github",
		GitHubID:     ghUser.ID,
		AvatarURL:    ghUser.AvatarURL,
		Plan:         "free",
	}

	// Upsert user in PostgreSQL (if DB is connected)
	if h.userStore != nil {
		if err := h.userStore.UpsertUser(c.Request().Context(), user); err != nil {
			slog.Error("auth.upsert_user.failed", "error", err, "github_id", ghUser.ID)
			// Non-fatal: login still works with in-memory user data
		} else {
			slog.Info("auth.upsert_user.success", "user_id", user.ID, "plan", user.Plan)
		}
	}

	// Generate tokens
	accessToken, err := h.jwt.GenerateAccessToken(user.ID, user.Plan, "", "user")
	if err != nil {
		return response.Err(c, err)
	}

	refreshToken, err := auth.GenerateRefreshToken()
	if err != nil {
		return response.Err(c, err)
	}

	// Store refresh token hash in-memory (replace with DB in prod)
	family := generateFamily() // H-04: token family for reuse detection
	h.mu.Lock()
	h.refresh[hashToken(refreshToken)] = refreshEntry{
		userID:    user.ID,
		plan:      user.Plan,
		tokenHash: hashToken(refreshToken),
		family:    family,
		expiresAt: time.Now().Add(auth.RefreshTokenTTL),
	}
	h.mu.Unlock()

	// L-04: Security audit logging
	slog.Info("auth.login.success", "user_id", user.ID, "provider", "github", "ip", c.RealIP())

	// If request came from dashboard, redirect back with a one-time auth code
	if entry.redirectTo == "dashboard" && h.dashboardURL != "" {
		code, codeErr := generateAuthCode()
		if codeErr != nil {
			return response.Err(c, codeErr)
		}
		h.mu.Lock()
		h.authCodes[code] = authCodeEntry{
			accessToken:  accessToken,
			refreshToken: refreshToken,
			expiresAt:    time.Now().Add(authCodeTTL),
		}
		h.mu.Unlock()
		redirectURL := fmt.Sprintf("%s/auth/callback?code=%s", h.dashboardURL, code)
		return c.Redirect(http.StatusTemporaryRedirect, redirectURL)
	}

	// If request came from CLI, redirect tokens to CLI's local callback server
	if len(entry.redirectTo) > 4 && entry.redirectTo[:4] == "cli:" {
		port := entry.redirectTo[4:]
		redirectURL := fmt.Sprintf("http://127.0.0.1:%s/callback?access_token=%s&refresh_token=%s&user_id=%s&plan=%s",
			port, accessToken, refreshToken, user.ID, user.Plan)
		return c.Redirect(http.StatusTemporaryRedirect, redirectURL)
	}

	// API-only flow: return JSON
	return response.OK(c, http.StatusOK, map[string]any{
		"user": user,
		"tokens": domain.TokenPair{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
			ExpiresIn:    int(auth.AccessTokenTTL.Seconds()),
		},
	})
}

// RefreshToken exchanges a refresh token for a new token pair (rotation).
func (h *AuthHandler) RefreshToken(c echo.Context) error {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := c.Bind(&req); err != nil || req.RefreshToken == "" {
		return response.ErrMsg(c, http.StatusBadRequest, "BAD_REQUEST", "refresh_token required")
	}

	tokenHash := hashToken(req.RefreshToken)

	h.mu.Lock()
	entry, ok := h.refresh[tokenHash]
	if ok {
		delete(h.refresh, tokenHash) // invalidate on use (rotation)
	}
	h.mu.Unlock()

	if !ok {
		// H-04: Potential token reuse — revoke all tokens in the family
		// (In production, this would query DB by family ID)
		return response.ErrMsg(c, http.StatusUnauthorized, "UNAUTHORIZED", "invalid refresh token")
	}
	if time.Now().After(entry.expiresAt) {
		return response.ErrMsg(c, http.StatusUnauthorized, "UNAUTHORIZED", "expired refresh token")
	}

	// Issue new token pair
	newAccess, err := h.jwt.GenerateAccessToken(entry.userID, entry.plan, "", "user")
	if err != nil {
		return response.Err(c, err)
	}

	newRefresh, err := auth.GenerateRefreshToken()
	if err != nil {
		return response.Err(c, err)
	}

	// Store rotated refresh token with same family (H-04)
	h.mu.Lock()
	h.refresh[hashToken(newRefresh)] = refreshEntry{
		userID:    entry.userID,
		plan:      entry.plan,
		tokenHash: hashToken(newRefresh),
		family:    entry.family, // H-04: preserve family
		expiresAt: time.Now().Add(auth.RefreshTokenTTL),
	}
	h.mu.Unlock()

	return response.OK(c, http.StatusOK, domain.TokenPair{
		AccessToken:  newAccess,
		RefreshToken: newRefresh,
		ExpiresIn:    int(auth.AccessTokenTTL.Seconds()),
	})
}

// Me returns the current authenticated user.
// When a user store is available, returns full profile (email, name, avatar_url).
// Falls back to JWT claims only when no user store is configured.
func (h *AuthHandler) Me(c echo.Context) error {
	claims := middleware.GetClaims(c)
	if claims == nil {
		return response.Err(c, domain.ErrUnauthorized)
	}

	// Try to fetch full user from DB if user store is available
	if h.userStore != nil {
		user, err := h.userStore.GetUserByID(c.Request().Context(), claims.UserID)
		if err == nil {
			return response.OK(c, http.StatusOK, map[string]any{
				"user_id":    user.ID,
				"email":      user.Email,
				"name":       user.Name,
				"avatar_url": user.AvatarURL,
				"plan":       user.Plan,
			})
		}
		// Fall through to claims-only response on error
		slog.Warn("auth.me.user_lookup_failed", "user_id", claims.UserID, "error", err)
	}

	return response.OK(c, http.StatusOK, map[string]string{
		"user_id": claims.UserID,
		"plan":    claims.Plan,
	})
}

// Signout revokes the refresh token (C-01: actual revocation).
func (h *AuthHandler) Signout(c echo.Context) error {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	// Bind is optional — signout should succeed even without body
	_ = c.Bind(&req)

	if req.RefreshToken != "" {
		tokenHash := hashToken(req.RefreshToken)
		h.mu.Lock()
		// H-04: Also revoke all tokens in the same family
		if entry, ok := h.refresh[tokenHash]; ok {
			family := entry.family
			for k, v := range h.refresh {
				if v.family == family {
					delete(h.refresh, k)
				}
			}
		}
		h.mu.Unlock()
	}

	return response.OK(c, http.StatusOK, map[string]string{
		"message": "signed out",
	})
}

func generateUserID(githubID int64) string {
	return "gh_" + itoa(githubID)
}

func itoa(n int64) string {
	return fmt.Sprintf("%d", n)
}

// Exchange exchanges a one-time auth code for tokens (dashboard flow).
// POST /api/v1/auth/exchange {code} -> {access_token, refresh_token, expires_in}
func (h *AuthHandler) Exchange(c echo.Context) error {
	var req struct {
		Code string `json:"code"`
	}
	if err := c.Bind(&req); err != nil || req.Code == "" {
		return response.ErrMsg(c, http.StatusBadRequest, "BAD_REQUEST", "code required")
	}

	h.mu.Lock()
	entry, ok := h.authCodes[req.Code]
	if ok {
		delete(h.authCodes, req.Code) // one-time use
	}
	h.mu.Unlock()

	if !ok || entry.used {
		return response.ErrMsg(c, http.StatusBadRequest, "INVALID_CODE", "invalid or expired auth code")
	}
	if time.Now().After(entry.expiresAt) {
		return response.ErrMsg(c, http.StatusBadRequest, "INVALID_CODE", "auth code expired")
	}

	slog.Info("auth.exchange.success", "ip", c.RealIP())

	return response.OK(c, http.StatusOK, domain.TokenPair{
		AccessToken:  entry.accessToken,
		RefreshToken: entry.refreshToken,
		ExpiresIn:    int(auth.AccessTokenTTL.Seconds()),
	})
}

// generateAuthCode creates a random auth code prefixed with "ac_".
func generateAuthCode() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate auth code: %w", err)
	}
	return "ac_" + hex.EncodeToString(b), nil
}

// generateFamily creates a unique token family identifier (H-04).
func generateFamily() string {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		// Fallback to timestamp-based ID if CSPRNG fails
		return hex.EncodeToString([]byte(fmt.Sprintf("%d", time.Now().UnixNano())))
	}
	return hex.EncodeToString(b)
}
