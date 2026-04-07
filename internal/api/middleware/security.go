package middleware

import (
	"github.com/labstack/echo/v4"
)

// SecurityHeaders returns middleware that sets security-related HTTP response headers.
func SecurityHeaders() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			h := c.Response().Header()
			h.Set("Strict-Transport-Security", "max-age=63072000; includeSubDomains; preload")
			h.Set("X-Frame-Options", "DENY")
			h.Set("X-Content-Type-Options", "nosniff")
			h.Set("Content-Security-Policy", "default-src 'none'; frame-ancestors 'none'")
			h.Set("Referrer-Policy", "strict-origin-when-cross-origin")
			return next(c)
		}
	}
}
