// Package middleware provides HTTP middleware for the Tene Cloud API.
package middleware

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/tomo-kay/tene/internal/api/response"
	"github.com/tomo-kay/tene/internal/auth"
)

const (
	// ContextKeyClaims is the key for JWT claims in echo context.
	ContextKeyClaims = "claims"
)

// JWTAuth returns middleware that validates Bearer JWT tokens.
func JWTAuth(jwtSvc *auth.JWTService) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			header := c.Request().Header.Get("Authorization")
			if header == "" {
				return response.ErrMsg(c, http.StatusUnauthorized, "UNAUTHORIZED", "missing authorization header")
			}

			parts := strings.SplitN(header, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
				return response.ErrMsg(c, http.StatusUnauthorized, "UNAUTHORIZED", "invalid authorization format")
			}

			claims, err := jwtSvc.ValidateAccessToken(parts[1])
			if err != nil {
				return response.ErrMsg(c, http.StatusUnauthorized, "UNAUTHORIZED", "invalid or expired token")
			}

			c.Set(ContextKeyClaims, claims)
			return next(c)
		}
	}
}

// GetClaims extracts JWT claims from the echo context.
func GetClaims(c echo.Context) *auth.Claims {
	claims, _ := c.Get(ContextKeyClaims).(*auth.Claims)
	return claims
}
