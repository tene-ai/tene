// Package auth provides authentication services for Tene Cloud.
package auth

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	AccessTokenTTL    = 15 * time.Minute
	RefreshTokenTTL   = 30 * 24 * time.Hour // 30 days
	MinJWTSecretLength = 32                  // C-02: minimum 256-bit secret for HS256
	JWTAudience       = "tene-api"           // H-03: audience claim
)

// Claims represents the JWT access token claims.
type Claims struct {
	UserID   string `json:"sub"`
	Plan     string `json:"plan"`
	DeviceID string `json:"did,omitempty"`
	Scope    string `json:"scope,omitempty"`
	jwt.RegisteredClaims
}

// JWTService handles JWT token generation and validation.
type JWTService struct {
	secret []byte
}

// NewJWTService creates a new JWT service with the given signing secret.
// Panics if the secret is shorter than 32 bytes (C-02).
func NewJWTService(secret string) *JWTService {
	if len(secret) < MinJWTSecretLength {
		panic(fmt.Sprintf("auth: JWT_SECRET must be at least %d characters, got %d", MinJWTSecretLength, len(secret)))
	}
	return &JWTService{secret: []byte(secret)}
}

// GenerateAccessToken creates a signed JWT access token.
func (s *JWTService) GenerateAccessToken(userID, plan, deviceID, scope string) (string, error) {
	now := time.Now()
	jti, err := generateJTI() // L-03: unique token ID for revocation
	if err != nil {
		return "", fmt.Errorf("auth: generate jti: %w", err)
	}

	claims := &Claims{
		UserID:   userID,
		Plan:     plan,
		DeviceID: deviceID,
		Scope:    scope,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        jti,                                    // L-03: jti for blacklisting
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(AccessTokenTTL)),
			Issuer:    "tene",
			Audience:  jwt.ClaimStrings{JWTAudience},          // H-03: audience claim
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(s.secret)
	if err != nil {
		return "", fmt.Errorf("auth: sign access token: %w", err)
	}
	return signed, nil
}

// ValidateAccessToken parses and validates a JWT access token.
func (s *JWTService) ValidateAccessToken(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("auth: unexpected signing method: %v", t.Header["alg"])
		}
		return s.secret, nil
	},
		jwt.WithAudience(JWTAudience), // H-03: validate audience
		jwt.WithIssuer("tene"),
	)
	if err != nil {
		return nil, fmt.Errorf("auth: parse token: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("auth: invalid token claims")
	}
	return claims, nil
}

// GenerateRefreshToken creates a cryptographically random refresh token.
func GenerateRefreshToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("auth: generate refresh token: %w", err)
	}
	return "rt_" + hex.EncodeToString(b), nil
}

// generateJTI creates a unique token identifier.
func generateJTI() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
