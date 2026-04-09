package domain

import "time"

// User represents a Tene Cloud user account.
type User struct {
	ID               string    `json:"id"`
	Email            string    `json:"email"`
	Name             string    `json:"name"`
	AuthProvider     string    `json:"auth_provider"` // "github" or "google"
	GitHubID         int64     `json:"github_id,omitempty"`
	GoogleID         string    `json:"google_id,omitempty"`
	LemonCustomerID  string    `json:"lemon_customer_id,omitempty"`
	X25519PublicKey  []byte    `json:"x25519_public_key,omitempty"`
	AvatarURL        string    `json:"avatar_url,omitempty"`
	Plan             string    `json:"plan"` // "free" or "pro"
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// TokenPair holds access and refresh tokens.
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"` // seconds
}
