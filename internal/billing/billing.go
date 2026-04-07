// Package billing provides LemonSqueezy integration for Tene Cloud subscriptions.
package billing

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Subscription represents a user's subscription state.
type Subscription struct {
	Plan      string     `json:"plan"`       // "free" or "pro"
	Status    string     `json:"status"`     // "active", "cancelled", "past_due", "expired"
	ExpiresAt *time.Time `json:"expires_at"` // current billing period end
}

// Config holds LemonSqueezy configuration.
type Config struct {
	APIKey         string // LEMONSQUEEZY_API_KEY
	WebhookSecret  string // LEMONSQUEEZY_WEBHOOK_SECRET
	StoreID        string // LEMONSQUEEZY_STORE_ID
	ProVariantID   string // LEMONSQUEEZY_VARIANT_PRO
	DashboardURL   string // DASHBOARD_URL for checkout redirect
}

// WebhookEvent represents a parsed LemonSqueezy webhook payload.
type WebhookEvent struct {
	Meta struct {
		EventName string `json:"event_name"`
	} `json:"meta"`
	Data struct {
		ID         string `json:"id"`
		Attributes struct {
			Status         string     `json:"status"`
			CustomerID     int64      `json:"customer_id"`
			VariantID      int64      `json:"variant_id"`
			UserEmail      string     `json:"user_email"`
			RenewsAt       *time.Time `json:"renews_at"`
			EndsAt         *time.Time `json:"ends_at"`
			SubscriptionID int64      `json:"first_subscription_item,omitempty"`
		} `json:"attributes"`
	} `json:"data"`
}

// UserStore defines user operations needed by billing.
type UserStore interface {
	UpdatePlan(ctx context.Context, email string, plan string, lemonCustomerID string) error
	GetLemonCustomerID(ctx context.Context, userID string) (string, error)
}

// Service handles LemonSqueezy billing operations.
type Service struct {
	cfg        Config
	httpClient *http.Client
	store      UserStore
}

// NewService creates a billing service.
func NewService(cfg Config, store UserStore) *Service {
	return &Service{
		cfg:        cfg,
		httpClient: &http.Client{Timeout: 10 * time.Second},
		store:      store,
	}
}

// CreateCheckoutURL generates a LemonSqueezy checkout URL for Pro upgrade.
func (s *Service) CreateCheckoutURL(ctx context.Context, userID, email string) (string, error) {
	if s.cfg.APIKey == "" || s.cfg.StoreID == "" || s.cfg.ProVariantID == "" {
		return "", fmt.Errorf("billing: LemonSqueezy not configured")
	}

	payload := map[string]any{
		"data": map[string]any{
			"type": "checkouts",
			"attributes": map[string]any{
				"checkout_data": map[string]any{
					"email":   email,
					"custom":  map[string]string{"user_id": userID},
				},
				"product_options": map[string]any{
					"redirect_url": s.cfg.DashboardURL + "/billing?success=true",
				},
			},
			"relationships": map[string]any{
				"store": map[string]any{
					"data": map[string]any{"type": "stores", "id": s.cfg.StoreID},
				},
				"variant": map[string]any{
					"data": map[string]any{"type": "variants", "id": s.cfg.ProVariantID},
				},
			},
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("billing: marshal checkout: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		"https://api.lemonsqueezy.com/v1/checkouts",
		bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+s.cfg.APIKey)
	req.Header.Set("Content-Type", "application/vnd.api+json")
	req.Header.Set("Accept", "application/vnd.api+json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("billing: checkout API: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("billing: read checkout response: %w", err)
	}

	if resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("billing: checkout API returned %d: %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		Data struct {
			Attributes struct {
				URL string `json:"url"`
			} `json:"attributes"`
		} `json:"data"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", fmt.Errorf("billing: decode checkout: %w", err)
	}

	return result.Data.Attributes.URL, nil
}

// GetPortalURL generates a LemonSqueezy customer portal URL.
func (s *Service) GetPortalURL(ctx context.Context, userID string) (string, error) {
	customerID, err := s.store.GetLemonCustomerID(ctx, userID)
	if err != nil || customerID == "" {
		return "", fmt.Errorf("billing: no subscription found. Upgrade first with 'tene billing upgrade'")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		fmt.Sprintf("https://api.lemonsqueezy.com/v1/customers/%s", customerID), nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+s.cfg.APIKey)
	req.Header.Set("Accept", "application/vnd.api+json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("billing: portal API: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	var result struct {
		Data struct {
			Attributes struct {
				URLs struct {
					CustomerPortal string `json:"customer_portal"`
				} `json:"urls"`
			} `json:"attributes"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("billing: decode portal: %w", err)
	}

	return result.Data.Attributes.URLs.CustomerPortal, nil
}

// VerifyWebhookSignature validates the HMAC SHA-256 signature from LemonSqueezy.
func (s *Service) VerifyWebhookSignature(payload []byte, signature string) bool {
	if s.cfg.WebhookSecret == "" {
		return false
	}
	mac := hmac.New(sha256.New, []byte(s.cfg.WebhookSecret))
	mac.Write(payload)
	expected := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(expected), []byte(signature))
}

// HandleWebhook processes a LemonSqueezy webhook event.
func (s *Service) HandleWebhook(ctx context.Context, payload []byte, signature string) error {
	if !s.VerifyWebhookSignature(payload, signature) {
		return fmt.Errorf("billing: invalid webhook signature")
	}

	var event WebhookEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		return fmt.Errorf("billing: parse webhook: %w", err)
	}

	email := event.Data.Attributes.UserEmail
	customerID := fmt.Sprintf("%d", event.Data.Attributes.CustomerID)

	switch event.Meta.EventName {
	case "subscription_created", "subscription_updated":
		if event.Data.Attributes.Status == "active" {
			return s.store.UpdatePlan(ctx, email, "pro", customerID)
		}
		if event.Data.Attributes.Status == "cancelled" || event.Data.Attributes.Status == "expired" {
			return s.store.UpdatePlan(ctx, email, "free", customerID)
		}
	case "subscription_cancelled":
		return s.store.UpdatePlan(ctx, email, "free", customerID)
	case "subscription_payment_failed":
		// 7-day grace period handled by LemonSqueezy; log for monitoring
		return nil
	}

	return nil
}

