package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tomo-kay/tene/internal/api/middleware"
	"github.com/tomo-kay/tene/internal/auth"
)

// setClaims is a test helper that sets JWT claims in the echo context.
func setClaims(c echo.Context, claims *auth.Claims) {
	c.Set(middleware.ContextKeyClaims, claims)
}

// parseResponse parses a JSON response body into the target struct.
func parseResponse(t *testing.T, rec *httptest.ResponseRecorder, target any) {
	t.Helper()
	err := json.Unmarshal(rec.Body.Bytes(), target)
	require.NoError(t, err, "failed to parse response body: %s", rec.Body.String())
}

func TestVaultHandler_Create(t *testing.T) {
	tests := []struct {
		name       string
		plan       string
		body       string
		wantStatus int
		wantError  string
	}{
		{
			name:       "pro plan success",
			plan:       "pro",
			body:       `{"project_name":"my-project"}`,
			wantStatus: http.StatusCreated,
		},
		{
			name:       "free plan rejected",
			plan:       "free",
			body:       `{"project_name":"my-project"}`,
			wantStatus: http.StatusPaymentRequired,
			wantError:  "PRO_PLAN_REQUIRED",
		},
		{
			name:       "duplicate project",
			plan:       "pro",
			body:       `{"project_name":"existing-project"}`,
			wantStatus: http.StatusConflict,
			wantError:  "PROJECT_EXISTS",
		},
		{
			name:       "missing project name",
			plan:       "pro",
			body:       `{}`,
			wantStatus: http.StatusBadRequest,
			wantError:  "BAD_REQUEST",
		},
		{
			name:       "invalid project name",
			plan:       "pro",
			body:       `{"project_name":"../traversal"}`,
			wantStatus: http.StatusBadRequest,
			wantError:  "BAD_REQUEST",
		},
		{
			name:       "no auth",
			plan:       "",
			body:       `{"project_name":"my-project"}`,
			wantStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := NewMemVaultStore()
			h := NewVaultHandler(store, NewMemVaultKeyMetadataStore(), nil)

			// Pre-create a vault for the duplicate test
			if tt.name == "duplicate project" {
				e := echo.New()
				req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`{"project_name":"existing-project"}`))
				req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
				rec := httptest.NewRecorder()
				c := e.NewContext(req, rec)
				setClaims(c, &auth.Claims{UserID: "user-1", Plan: "pro"})
				require.NoError(t, h.Create(c))
			}

			e := echo.New()
			req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(tt.body))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			if tt.plan != "" {
				setClaims(c, &auth.Claims{UserID: "user-1", Plan: tt.plan})
			}

			err := h.Create(c)
			require.NoError(t, err)
			assert.Equal(t, tt.wantStatus, rec.Code)

			if tt.wantError != "" {
				var resp struct {
					OK    bool   `json:"ok"`
					Error string `json:"error"`
				}
				parseResponse(t, rec, &resp)
				assert.False(t, resp.OK)
				assert.Equal(t, tt.wantError, resp.Error)
			}
		})
	}
}

func TestVaultHandler_List(t *testing.T) {
	tests := []struct {
		name       string
		setupUser  string
		queryUser  string
		plan       string
		seedCount  int
		wantStatus int
		wantCount  int
	}{
		{
			name:       "list own vaults",
			setupUser:  "user-1",
			queryUser:  "user-1",
			plan:       "pro",
			seedCount:  2,
			wantStatus: http.StatusOK,
			wantCount:  2,
		},
		{
			name:       "empty list",
			setupUser:  "user-1",
			queryUser:  "user-2",
			plan:       "free",
			seedCount:  0,
			wantStatus: http.StatusOK,
			wantCount:  0,
		},
		{
			name:       "no auth",
			queryUser:  "",
			plan:       "",
			wantStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := NewMemVaultStore()
			h := NewVaultHandler(store, NewMemVaultKeyMetadataStore(), nil)

			// Seed vaults
			for i := 0; i < tt.seedCount; i++ {
				e := echo.New()
				body := `{"project_name":"project-` + string(rune('a'+i)) + `"}`
				req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(body))
				req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
				rec := httptest.NewRecorder()
				c := e.NewContext(req, rec)
				setClaims(c, &auth.Claims{UserID: tt.setupUser, Plan: "pro"})
				require.NoError(t, h.Create(c))
			}

			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			if tt.plan != "" {
				setClaims(c, &auth.Claims{UserID: tt.queryUser, Plan: tt.plan})
			}

			err := h.List(c)
			require.NoError(t, err)
			assert.Equal(t, tt.wantStatus, rec.Code)

			if tt.wantStatus == http.StatusOK {
				var resp struct {
					OK   bool `json:"ok"`
					Data []struct {
						ID string `json:"id"`
					} `json:"data"`
				}
				parseResponse(t, rec, &resp)
				assert.True(t, resp.OK)
				assert.Len(t, resp.Data, tt.wantCount)
			}
		})
	}
}

func TestVaultHandler_Get(t *testing.T) {
	store := NewMemVaultStore()
	h := NewVaultHandler(store, NewMemVaultKeyMetadataStore(), nil)

	// Create a vault first
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`{"project_name":"test-project"}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	setClaims(c, &auth.Claims{UserID: "user-1", Plan: "pro"})
	require.NoError(t, h.Create(c))

	var createResp struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	parseResponse(t, rec, &createResp)
	vaultID := createResp.Data.ID

	tests := []struct {
		name       string
		vaultID    string
		userID     string
		plan       string
		wantStatus int
	}{
		{
			name:       "get own vault",
			vaultID:    vaultID,
			userID:     "user-1",
			plan:       "pro",
			wantStatus: http.StatusOK,
		},
		{
			name:       "free plan rejected",
			vaultID:    vaultID,
			userID:     "user-1",
			plan:       "free",
			wantStatus: http.StatusPaymentRequired,
		},
		{
			name:       "vault not found",
			vaultID:    "nonexistent",
			userID:     "user-1",
			plan:       "pro",
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "wrong user",
			vaultID:    vaultID,
			userID:     "user-2",
			plan:       "pro",
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetParamNames("id")
			c.SetParamValues(tt.vaultID)
			setClaims(c, &auth.Claims{UserID: tt.userID, Plan: tt.plan})

			err := h.Get(c)
			require.NoError(t, err)
			assert.Equal(t, tt.wantStatus, rec.Code)
		})
	}
}

func TestVaultHandler_Delete(t *testing.T) {
	tests := []struct {
		name       string
		userID     string
		wantStatus int
	}{
		{
			name:       "delete own vault",
			userID:     "user-1",
			wantStatus: http.StatusOK,
		},
		{
			name:       "delete other user vault",
			userID:     "user-2",
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := NewMemVaultStore()
			h := NewVaultHandler(store, NewMemVaultKeyMetadataStore(), nil)

			// Create a vault
			e := echo.New()
			req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`{"project_name":"to-delete"}`))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			setClaims(c, &auth.Claims{UserID: "user-1", Plan: "pro"})
			require.NoError(t, h.Create(c))

			var createResp struct {
				Data struct {
					ID string `json:"id"`
				} `json:"data"`
			}
			parseResponse(t, rec, &createResp)

			// Delete
			e = echo.New()
			req = httptest.NewRequest(http.MethodDelete, "/", nil)
			rec = httptest.NewRecorder()
			c = e.NewContext(req, rec)
			c.SetParamNames("id")
			c.SetParamValues(createResp.Data.ID)
			setClaims(c, &auth.Claims{UserID: tt.userID, Plan: "pro"})

			// S3 client is nil, so delete will log error but continue
			err := h.Delete(c)
			require.NoError(t, err)
			assert.Equal(t, tt.wantStatus, rec.Code)
		})
	}
}

func TestVaultHandler_Push_EmptyBody(t *testing.T) {
	store := NewMemVaultStore()
	h := NewVaultHandler(store, NewMemVaultKeyMetadataStore(), nil)

	// Create vault first
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`{"project_name":"push-test"}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	setClaims(c, &auth.Claims{UserID: "user-1", Plan: "pro"})
	require.NoError(t, h.Create(c))

	var createResp struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	parseResponse(t, rec, &createResp)

	// Push with empty body
	e = echo.New()
	req = httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(""))
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(createResp.Data.ID)
	setClaims(c, &auth.Claims{UserID: "user-1", Plan: "pro"})

	err := h.Push(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var resp struct {
		Error string `json:"error"`
	}
	parseResponse(t, rec, &resp)
	assert.Equal(t, "BAD_REQUEST", resp.Error)
}

func TestVaultHandler_Push_FreePlanRejected(t *testing.T) {
	store := NewMemVaultStore()
	h := NewVaultHandler(store, NewMemVaultKeyMetadataStore(), nil)

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString("data"))
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("some-id")
	setClaims(c, &auth.Claims{UserID: "user-1", Plan: "free"})

	err := h.Push(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusPaymentRequired, rec.Code)
}

func TestVaultHandler_Pull_NoPush(t *testing.T) {
	store := NewMemVaultStore()
	h := NewVaultHandler(store, NewMemVaultKeyMetadataStore(), nil)

	// Create vault (version 0, never pushed)
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`{"project_name":"pull-test"}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	setClaims(c, &auth.Claims{UserID: "user-1", Plan: "pro"})
	require.NoError(t, h.Create(c))

	var createResp struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	parseResponse(t, rec, &createResp)

	// Pull from never-pushed vault
	e = echo.New()
	req = httptest.NewRequest(http.MethodGet, "/", nil)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(createResp.Data.ID)
	setClaims(c, &auth.Claims{UserID: "user-1", Plan: "pro"})

	err := h.Pull(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, rec.Code)

	var resp struct {
		Error string `json:"error"`
	}
	parseResponse(t, rec, &resp)
	assert.Equal(t, "VAULT_EMPTY", resp.Error)
}
