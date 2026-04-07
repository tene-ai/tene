package billing

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockUserStore struct {
	plans map[string]string // email -> plan
}

func (m *mockUserStore) UpdatePlan(_ context.Context, email, plan, customerID string) error {
	m.plans[email] = plan
	return nil
}

func (m *mockUserStore) GetLemonCustomerID(_ context.Context, userID string) (string, error) {
	return "", nil
}

func TestVerifyWebhookSignature_Valid(t *testing.T) {
	secret := "test-webhook-secret"
	svc := NewService(Config{WebhookSecret: secret}, nil)

	payload := []byte(`{"test": true}`)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	signature := hex.EncodeToString(mac.Sum(nil))

	assert.True(t, svc.VerifyWebhookSignature(payload, signature))
}

func TestVerifyWebhookSignature_Invalid(t *testing.T) {
	svc := NewService(Config{WebhookSecret: "real-secret"}, nil)

	payload := []byte(`{"test": true}`)
	assert.False(t, svc.VerifyWebhookSignature(payload, "bad-signature"))
}

func TestVerifyWebhookSignature_EmptySecret(t *testing.T) {
	svc := NewService(Config{WebhookSecret: ""}, nil)

	assert.False(t, svc.VerifyWebhookSignature([]byte("data"), "any"))
}

func TestHandleWebhook_SubscriptionCreated(t *testing.T) {
	secret := "test-secret"
	store := &mockUserStore{plans: make(map[string]string)}
	svc := NewService(Config{WebhookSecret: secret}, store)

	event := WebhookEvent{}
	event.Meta.EventName = "subscription_created"
	event.Data.Attributes.Status = "active"
	event.Data.Attributes.UserEmail = "alice@example.com"
	event.Data.Attributes.CustomerID = 12345

	payload, err := json.Marshal(event)
	require.NoError(t, err)

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	signature := hex.EncodeToString(mac.Sum(nil))

	err = svc.HandleWebhook(context.Background(), payload, signature)
	require.NoError(t, err)

	assert.Equal(t, "pro", store.plans["alice@example.com"])
}

func TestHandleWebhook_SubscriptionCancelled(t *testing.T) {
	secret := "test-secret"
	store := &mockUserStore{plans: map[string]string{"bob@example.com": "pro"}}
	svc := NewService(Config{WebhookSecret: secret}, store)

	event := WebhookEvent{}
	event.Meta.EventName = "subscription_cancelled"
	event.Data.Attributes.UserEmail = "bob@example.com"
	event.Data.Attributes.CustomerID = 67890

	payload, err := json.Marshal(event)
	require.NoError(t, err)

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	signature := hex.EncodeToString(mac.Sum(nil))

	err = svc.HandleWebhook(context.Background(), payload, signature)
	require.NoError(t, err)

	assert.Equal(t, "free", store.plans["bob@example.com"])
}

func TestHandleWebhook_InvalidSignature(t *testing.T) {
	svc := NewService(Config{WebhookSecret: "real-secret"}, nil)

	err := svc.HandleWebhook(context.Background(), []byte(`{}`), "wrong-sig")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid webhook signature")
}

func TestCreateCheckoutURL_NotConfigured(t *testing.T) {
	svc := NewService(Config{}, nil)

	_, err := svc.CreateCheckoutURL(context.Background(), "user-1", "user@test.com")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not configured")
}
