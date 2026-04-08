// Package handler provides HTTP handlers for the Tene Cloud API.
package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

// DBPinger is implemented by any DB pool that supports ping for health checks.
type DBPinger interface {
	Ping(ctx context.Context) error
}

// HealthHandler handles liveness and readiness probes.
type HealthHandler struct {
	DB DBPinger // optional: nil skips DB check
}

// Liveness returns 200 OK if the process is alive (ECS health check).
func (h *HealthHandler) Liveness(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
}

// Readiness returns 200 OK if the service is ready to handle requests.
// When a DB is configured, it also pings the database.
func (h *HealthHandler) Readiness(c echo.Context) error {
	if h.DB != nil {
		ctx, cancel := context.WithTimeout(c.Request().Context(), 3*time.Second)
		defer cancel()
		if err := h.DB.Ping(ctx); err != nil {
			return c.JSON(http.StatusServiceUnavailable, map[string]string{
				"status": "not ready",
				"reason": "database unreachable",
			})
		}
	}
	return c.JSON(http.StatusOK, map[string]string{"status": "ready"})
}
