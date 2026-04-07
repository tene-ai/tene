package handler

import (
	"net/http"
	"sync"

	"github.com/labstack/echo/v4"
	"github.com/tomo-kay/tene/internal/api/middleware"
	"github.com/tomo-kay/tene/internal/api/response"
	"github.com/tomo-kay/tene/internal/domain"
)

// AuditStore defines database operations for audit logs.
type AuditStore interface {
	ListAuditLogs(userID string, limit int) ([]domain.AuditLog, error)
}

// MemAuditStore is an in-memory AuditStore for development.
type MemAuditStore struct {
	mu   sync.RWMutex
	logs []domain.AuditLog
}

// NewMemAuditStore creates an in-memory audit store.
func NewMemAuditStore() *MemAuditStore {
	return &MemAuditStore{}
}

// ListAuditLogs returns recent audit events for a user.
func (s *MemAuditStore) ListAuditLogs(userID string, limit int) ([]domain.AuditLog, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var result []domain.AuditLog
	for i := len(s.logs) - 1; i >= 0 && len(result) < limit; i-- {
		if s.logs[i].UserID == userID {
			result = append(result, s.logs[i])
		}
	}
	return result, nil
}

// AuditHandler handles audit log queries.
type AuditHandler struct {
	store AuditStore
}

// NewAuditHandler creates an audit handler.
func NewAuditHandler(store AuditStore) *AuditHandler {
	return &AuditHandler{store: store}
}

// List returns audit events for the authenticated user.
func (h *AuditHandler) List(c echo.Context) error {
	claims := middleware.GetClaims(c)
	if claims == nil {
		return response.Err(c, domain.ErrUnauthorized)
	}

	logs, err := h.store.ListAuditLogs(claims.UserID, 100)
	if err != nil {
		return response.Err(c, err)
	}
	return response.OK(c, http.StatusOK, logs)
}
