package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tomo-kay/tene/internal/auth"
)

// newTestAuthHandler creates an AuthHandler suitable for unit tests.
// Uses a real JWTService with a test secret (no OAuth calls needed).
func newTestAuthHandler() *AuthHandler {
	jwtSvc := auth.NewJWTService("test-secret-must-be-at-least-32-bytes-long")
	oauthSvc := auth.NewOAuthService("test-client-id", "test-client-secret", "http://localhost:8080")
	return NewAuthHandler(oauthSvc, jwtSvc, "http://localhost:3001")
}

func TestAuthHandler_Exchange(t *testing.T) {
	h := newTestAuthHandler()

	// Seed a valid auth code
	code := "ac_test123456789"
	jwtSvc := auth.NewJWTService("test-secret-must-be-at-least-32-bytes-long")
	accessToken, err := jwtSvc.GenerateAccessToken("user-1", "pro", "", "user")
	require.NoError(t, err)
	refreshToken, err := auth.GenerateRefreshToken()
	require.NoError(t, err)

	h.mu.Lock()
	h.authCodes[code] = authCodeEntry{
		accessToken:  accessToken,
		refreshToken: refreshToken,
		expiresAt:    h.authCodes[code].expiresAt, // zero time = already expired
	}
	h.mu.Unlock()

	tests := []struct {
		name       string
		code       string
		wantStatus int
		wantError  string
	}{
		{
			name:       "invalid code",
			code:       "ac_nonexistent",
			wantStatus: http.StatusBadRequest,
			wantError:  "INVALID_CODE",
		},
		{
			name:       "missing code",
			code:       "",
			wantStatus: http.StatusBadRequest,
			wantError:  "BAD_REQUEST",
		},
		{
			name:       "expired code",
			code:       code, // expiresAt is zero time, so it's expired
			wantStatus: http.StatusBadRequest,
			wantError:  "INVALID_CODE",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			body, _ := json.Marshal(map[string]string{"code": tt.code})
			req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			err := h.Exchange(c)
			require.NoError(t, err)
			assert.Equal(t, tt.wantStatus, rec.Code)

			if tt.wantError != "" {
				var resp struct {
					Error string `json:"error"`
				}
				parseResponse(t, rec, &resp)
				assert.Equal(t, tt.wantError, resp.Error)
			}
		})
	}
}

func TestAuthHandler_Exchange_ValidCode(t *testing.T) {
	h := newTestAuthHandler()

	jwtSvc := auth.NewJWTService("test-secret-must-be-at-least-32-bytes-long")
	accessToken, err := jwtSvc.GenerateAccessToken("user-1", "pro", "", "user")
	require.NoError(t, err)
	refreshToken, err := auth.GenerateRefreshToken()
	require.NoError(t, err)

	code := "ac_valid_code_12345"
	h.mu.Lock()
	h.authCodes[code] = authCodeEntry{
		accessToken:  accessToken,
		refreshToken: refreshToken,
		expiresAt:    time.Now().Add(5 * time.Minute),
	}
	h.mu.Unlock()

	e := echo.New()
	body, _ := json.Marshal(map[string]string{"code": code})
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err = h.Exchange(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var resp struct {
		OK   bool `json:"ok"`
		Data struct {
			AccessToken  string `json:"access_token"`
			RefreshToken string `json:"refresh_token"`
			ExpiresIn    int    `json:"expires_in"`
		} `json:"data"`
	}
	parseResponse(t, rec, &resp)
	assert.True(t, resp.OK)
	assert.NotEmpty(t, resp.Data.AccessToken)
	assert.NotEmpty(t, resp.Data.RefreshToken)
	assert.Greater(t, resp.Data.ExpiresIn, 0)

	// Code should be consumed (one-time use)
	e = echo.New()
	body, _ = json.Marshal(map[string]string{"code": code})
	req = httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	err = h.Exchange(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestAuthHandler_Me(t *testing.T) {
	tests := []struct {
		name       string
		claims     *auth.Claims
		wantStatus int
	}{
		{
			name:       "authenticated user",
			claims:     &auth.Claims{UserID: "user-1", Plan: "pro"},
			wantStatus: http.StatusOK,
		},
		{
			name:       "no auth",
			claims:     nil,
			wantStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newTestAuthHandler()

			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			if tt.claims != nil {
				setClaims(c, tt.claims)
			}

			err := h.Me(c)
			require.NoError(t, err)
			assert.Equal(t, tt.wantStatus, rec.Code)

			if tt.wantStatus == http.StatusOK {
				var resp struct {
					OK   bool `json:"ok"`
					Data struct {
						UserID string `json:"user_id"`
						Plan   string `json:"plan"`
					} `json:"data"`
				}
				parseResponse(t, rec, &resp)
				assert.True(t, resp.OK)
				assert.Equal(t, tt.claims.UserID, resp.Data.UserID)
				assert.Equal(t, tt.claims.Plan, resp.Data.Plan)
			}
		})
	}
}

func TestAuthHandler_RefreshToken(t *testing.T) {
	h := newTestAuthHandler()

	// Generate a refresh token and store it
	refreshToken, err := auth.GenerateRefreshToken()
	require.NoError(t, err)

	h.mu.Lock()
	h.refresh[hashToken(refreshToken)] = refreshEntry{
		userID:    "user-1",
		plan:      "pro",
		tokenHash: hashToken(refreshToken),
		family:    "fam-1",
		expiresAt: time.Now().Add(24 * time.Hour),
	}
	h.mu.Unlock()

	tests := []struct {
		name       string
		token      string
		wantStatus int
	}{
		{
			name:       "valid refresh",
			token:      refreshToken,
			wantStatus: http.StatusOK,
		},
		{
			name:       "invalid refresh token",
			token:      "rt_invalid",
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "missing refresh token",
			token:      "",
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			body, _ := json.Marshal(map[string]string{"refresh_token": tt.token})
			req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			err := h.RefreshToken(c)
			require.NoError(t, err)
			assert.Equal(t, tt.wantStatus, rec.Code)
		})
	}
}

func TestAuthHandler_Signout(t *testing.T) {
	h := newTestAuthHandler()

	// Generate and store a refresh token
	refreshToken, err := auth.GenerateRefreshToken()
	require.NoError(t, err)

	h.mu.Lock()
	h.refresh[hashToken(refreshToken)] = refreshEntry{
		userID:    "user-1",
		plan:      "pro",
		tokenHash: hashToken(refreshToken),
		family:    "fam-signout",
		expiresAt: h.refresh[hashToken(refreshToken)].expiresAt,
	}
	h.mu.Unlock()

	e := echo.New()
	body, _ := json.Marshal(map[string]string{"refresh_token": refreshToken})
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	setClaims(c, &auth.Claims{UserID: "user-1", Plan: "pro"})

	err = h.Signout(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	// Verify token was revoked
	h.mu.RLock()
	_, exists := h.refresh[hashToken(refreshToken)]
	h.mu.RUnlock()
	assert.False(t, exists, "refresh token should be revoked after signout")
}
