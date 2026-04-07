package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testSecret = "test-secret-32-bytes-long-enough!" // >= MinJWTSecretLength

func TestJWT_GenerateAndValidate(t *testing.T) {
	svc := NewJWTService(testSecret)

	token, err := svc.GenerateAccessToken("user-123", "pro", "device-1", "user")
	require.NoError(t, err)
	assert.NotEmpty(t, token)

	claims, err := svc.ValidateAccessToken(token)
	require.NoError(t, err)
	assert.Equal(t, "user-123", claims.UserID)
	assert.Equal(t, "pro", claims.Plan)
	assert.Equal(t, "device-1", claims.DeviceID)
	assert.Equal(t, "user", claims.Scope)
	assert.Equal(t, "tene", claims.Issuer)
	assert.Contains(t, claims.Audience, JWTAudience) // H-03: audience validated
	assert.NotEmpty(t, claims.ID)                     // L-03: jti present
	assert.WithinDuration(t, time.Now().Add(AccessTokenTTL), claims.ExpiresAt.Time, 5*time.Second)
}

func TestJWT_InvalidToken(t *testing.T) {
	svc := NewJWTService(testSecret)

	_, err := svc.ValidateAccessToken("invalid.token.here")
	assert.Error(t, err)
}

func TestJWT_WrongSecret(t *testing.T) {
	svc1 := NewJWTService("secret-one-at-least-32-characters!")
	svc2 := NewJWTService("secret-two-at-least-32-characters!")

	token, err := svc1.GenerateAccessToken("user-1", "free", "", "user")
	require.NoError(t, err)

	_, err = svc2.ValidateAccessToken(token)
	assert.Error(t, err)
}

func TestJWT_FreePlan(t *testing.T) {
	svc := NewJWTService(testSecret)

	token, err := svc.GenerateAccessToken("user-456", "free", "", "user")
	require.NoError(t, err)

	claims, err := svc.ValidateAccessToken(token)
	require.NoError(t, err)
	assert.Equal(t, "free", claims.Plan)
	assert.Equal(t, "user", claims.Scope)
}

func TestJWT_ExpiredToken(t *testing.T) {
	svc := NewJWTService(testSecret)

	// Create token with past expiry
	claims := &Claims{
		UserID: "user-expired",
		Plan:   "free",
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)),
			Issuer:    "tene",
			Audience:  jwt.ClaimStrings{JWTAudience},
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	expiredToken, err := token.SignedString(svc.secret)
	require.NoError(t, err)

	_, err = svc.ValidateAccessToken(expiredToken)
	assert.Error(t, err, "expired token should fail validation")
}

func TestJWT_ShortSecretPanics(t *testing.T) {
	// C-02: short secrets should panic
	assert.Panics(t, func() {
		NewJWTService("too-short")
	})
}

func TestGenerateRefreshToken(t *testing.T) {
	token, err := GenerateRefreshToken()
	require.NoError(t, err)
	assert.True(t, len(token) > 10)
	assert.Contains(t, token, "rt_")

	token2, err := GenerateRefreshToken()
	require.NoError(t, err)
	assert.NotEqual(t, token, token2)
}
