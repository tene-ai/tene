package handler

import (
	"io"
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/tomo-kay/tene/internal/api/middleware"
	"github.com/tomo-kay/tene/internal/api/response"
	"github.com/tomo-kay/tene/internal/billing"
	"github.com/tomo-kay/tene/internal/domain"
)

// BillingHandler handles subscription and checkout endpoints.
type BillingHandler struct {
	svc *billing.Service
}

// NewBillingHandler creates a billing handler.
func NewBillingHandler(svc *billing.Service) *BillingHandler {
	return &BillingHandler{svc: svc}
}

// GetSubscription returns the current user's subscription status.
func (h *BillingHandler) GetSubscription(c echo.Context) error {
	claims := middleware.GetClaims(c)
	if claims == nil {
		return response.Err(c, domain.ErrUnauthorized)
	}

	// Plan is already in the JWT claims
	sub := billing.Subscription{
		Plan:   claims.Plan,
		Status: "active",
	}
	if claims.Plan == "free" {
		sub.Status = "none"
	}

	return response.OK(c, http.StatusOK, sub)
}

// CreateCheckout generates a LemonSqueezy checkout URL for Pro upgrade.
func (h *BillingHandler) CreateCheckout(c echo.Context) error {
	claims := middleware.GetClaims(c)
	if claims == nil {
		return response.Err(c, domain.ErrUnauthorized)
	}

	if claims.Plan == "pro" {
		return response.ErrMsg(c, http.StatusBadRequest, "ALREADY_PRO", "already on Pro plan")
	}

	var req struct {
		Email string `json:"email"`
	}
	if err := c.Bind(&req); err != nil || req.Email == "" {
		return response.ErrMsg(c, http.StatusBadRequest, "BAD_REQUEST", "email required")
	}

	url, err := h.svc.CreateCheckoutURL(c.Request().Context(), claims.UserID, req.Email)
	if err != nil {
		slog.Error("billing.checkout.failed", "user_id", claims.UserID, "error", err)
		return response.ErrMsg(c, http.StatusInternalServerError, "CHECKOUT_FAILED", "failed to create checkout")
	}

	return response.OK(c, http.StatusOK, map[string]string{
		"checkout_url": url,
	})
}

// CreatePortal generates a LemonSqueezy customer portal URL.
func (h *BillingHandler) CreatePortal(c echo.Context) error {
	claims := middleware.GetClaims(c)
	if claims == nil {
		return response.Err(c, domain.ErrUnauthorized)
	}

	url, err := h.svc.GetPortalURL(c.Request().Context(), claims.UserID)
	if err != nil {
		return response.ErrMsg(c, http.StatusBadRequest, "NO_SUBSCRIPTION", "no active subscription")
	}

	return response.OK(c, http.StatusOK, map[string]string{
		"portal_url": url,
	})
}

// Webhook handles incoming LemonSqueezy webhook events.
// This endpoint is public (no JWT) but validated with HMAC signature.
func (h *BillingHandler) Webhook(c echo.Context) error {
	signature := c.Request().Header.Get("X-Signature")
	if signature == "" {
		return response.ErrMsg(c, http.StatusBadRequest, "BAD_REQUEST", "missing signature")
	}

	body, err := io.ReadAll(io.LimitReader(c.Request().Body, 1<<20)) // 1MB limit
	if err != nil {
		return response.ErrMsg(c, http.StatusBadRequest, "BAD_REQUEST", "failed to read body")
	}

	if err := h.svc.HandleWebhook(c.Request().Context(), body, signature); err != nil {
		slog.Error("billing.webhook.failed", "error", err)
		return response.ErrMsg(c, http.StatusBadRequest, "WEBHOOK_FAILED", "webhook processing failed")
	}

	slog.Info("billing.webhook.processed", "ip", c.RealIP())
	return response.OK(c, http.StatusOK, map[string]string{"message": "ok"})
}
