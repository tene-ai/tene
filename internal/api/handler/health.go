// Package handler provides HTTP handlers for the Tene Cloud API.
package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// HealthHandler handles liveness and readiness probes.
type HealthHandler struct{}

// Liveness returns 200 OK if the process is alive (ECS health check).
func (h *HealthHandler) Liveness(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
}

// Readiness returns 200 OK if the service is ready to handle requests.
// TODO: Add DB ping when PostgreSQL is connected.
func (h *HealthHandler) Readiness(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{"status": "ready"})
}
