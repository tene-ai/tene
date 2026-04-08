package handler

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tomo-kay/tene/internal/auth"
	"github.com/tomo-kay/tene/internal/billing"
)

func TestBillingHandler_GetSubscription(t *testing.T) {
	tests := []struct {
		name       string
		claims     *auth.Claims
		wantStatus int
		wantPlan   string
		wantSubStatus string
	}{
		{
			name:          "pro plan user",
			claims:        &auth.Claims{UserID: "user-1", Plan: "pro"},
			wantStatus:    http.StatusOK,
			wantPlan:      "pro",
			wantSubStatus: "active",
		},
		{
			name:          "free plan user",
			claims:        &auth.Claims{UserID: "user-2", Plan: "free"},
			wantStatus:    http.StatusOK,
			wantPlan:      "free",
			wantSubStatus: "none",
		},
		{
			name:       "no auth",
			claims:     nil,
			wantStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := billing.NewService(billing.Config{}, nil)
			h := NewBillingHandler(svc)

			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			if tt.claims != nil {
				setClaims(c, tt.claims)
			}

			err := h.GetSubscription(c)
			require.NoError(t, err)
			assert.Equal(t, tt.wantStatus, rec.Code)

			if tt.wantStatus == http.StatusOK {
				var resp struct {
					OK   bool `json:"ok"`
					Data struct {
						Plan   string `json:"plan"`
						Status string `json:"status"`
					} `json:"data"`
				}
				parseResponse(t, rec, &resp)
				assert.True(t, resp.OK)
				assert.Equal(t, tt.wantPlan, resp.Data.Plan)
				assert.Equal(t, tt.wantSubStatus, resp.Data.Status)
			}
		})
	}
}

func TestBillingHandler_Webhook_MissingSignature(t *testing.T) {
	svc := billing.NewService(billing.Config{WebhookSecret: "test-secret"}, nil)
	h := NewBillingHandler(svc)

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.Webhook(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var resp struct {
		Error string `json:"error"`
	}
	parseResponse(t, rec, &resp)
	assert.Equal(t, "BAD_REQUEST", resp.Error)
}

func TestBillingHandler_Webhook_InvalidSignature(t *testing.T) {
	svc := billing.NewService(billing.Config{WebhookSecret: "test-secret"}, nil)
	h := NewBillingHandler(svc)

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`{"meta":{}}`))
	req.Header.Set("X-Signature", "invalid-sig")
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.Webhook(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var resp struct {
		Error string `json:"error"`
	}
	parseResponse(t, rec, &resp)
	assert.Equal(t, "WEBHOOK_FAILED", resp.Error)
}
