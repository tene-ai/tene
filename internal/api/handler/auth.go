package handler

import (
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

const stateTTL = 5 * time.Minute

type stateEntry struct {
	codeVerifier string
	expiresAt    time.Time
}

// refreshEntry stores metadata for a refresh token.
type refreshEntry struct {
	userID    string
	plan      string
	tokenHash string
	family    string    // H-04: token family ID for reuse detection
	expiresAt time.Time
}

// AuthHandler handles OAuth and token endpoints.
type AuthHandler struct {
	oauth    *auth.OAuthService
	jwt      *auth.JWTService
	mu       sync.RWMutex
	states   map[string]stateEntry   // in-memory state store with TTL (replace with Redis in prod)
	refresh  map[string]refreshEntry // in-memory refresh token store (replace with DB in prod)
}

// NewAuthHandler creates an auth handler.
func NewAuthHandler(oauth *auth.OAuthService, jwt *auth.JWTService) *AuthHandler {
	h := &AuthHandler{
		oauth:   oauth,
		jwt:     jwt,
		states:  make(map[string]stateEntry),
		refresh: make(map[string]refreshEntry),
	}
	go h.cleanupLoop()
	return h
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
	h.mu.Lock()
	h.states[state] = stateEntry{
		codeVerifier: pkce.CodeVerifier,
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

	// TODO: Upsert user in PostgreSQL when DB is connected
	// For now, create a user object from GitHub data
	user := &domain.User{
		ID:           generateUserID(ghUser.ID),
		Email:        ghUser.Email,
		Name:         ghUser.Name,
		AuthProvider: "github",
		GitHubID:     ghUser.ID,
		AvatarURL:    ghUser.AvatarURL,
		Plan:         "free",
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
func (h *AuthHandler) Me(c echo.Context) error {
	claims := middleware.GetClaims(c)
	if claims == nil {
		return response.Err(c, domain.ErrUnauthorized)
	}

	// TODO: Fetch full user from DB
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

// generateFamily creates a unique token family identifier (H-04).
func generateFamily() string {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		// Fallback to timestamp-based ID if CSPRNG fails
		return hex.EncodeToString([]byte(fmt.Sprintf("%d", time.Now().UnixNano())))
	}
	return hex.EncodeToString(b)
}
