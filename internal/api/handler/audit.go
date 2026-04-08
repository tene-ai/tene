package handler

import (
	"net/http"
	"strconv"
	"sync"

	"github.com/labstack/echo/v4"
	"github.com/tomo-kay/tene/internal/api/middleware"
	"github.com/tomo-kay/tene/internal/api/response"
	"github.com/tomo-kay/tene/internal/domain"
)

const (
	auditDefaultLimit = 100
	auditMaxLimit     = 500
)

// AuditStore defines database operations for audit logs.
type AuditStore interface {
	ListAuditLogs(userID string, filter domain.AuditFilter) ([]domain.AuditLog, error)
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

// ListAuditLogs returns recent audit events for a user with optional filtering.
func (s *MemAuditStore) ListAuditLogs(userID string, filter domain.AuditFilter) ([]domain.AuditLog, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []domain.AuditLog
	skipped := 0
	for i := len(s.logs) - 1; i >= 0 && len(result) < filter.Limit; i-- {
		if s.logs[i].UserID != userID {
			continue
		}
		if filter.Action != "" && s.logs[i].Action != filter.Action {
			continue
		}
		if skipped < filter.Offset {
			skipped++
			continue
		}
		result = append(result, s.logs[i])
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
// Supports query parameters: ?action=push&limit=50&offset=0
func (h *AuditHandler) List(c echo.Context) error {
	claims := middleware.GetClaims(c)
	if claims == nil {
		return response.Err(c, domain.ErrUnauthorized)
	}

	filter := domain.AuditFilter{
		Action: c.QueryParam("action"),
		Limit:  auditDefaultLimit,
		Offset: 0,
	}

	if v := c.QueryParam("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			filter.Limit = n
			if filter.Limit > auditMaxLimit {
				filter.Limit = auditMaxLimit
			}
		}
	}
	if v := c.QueryParam("offset"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			filter.Offset = n
		}
	}

	logs, err := h.store.ListAuditLogs(claims.UserID, filter)
	if err != nil {
		return response.Err(c, err)
	}
	return response.OK(c, http.StatusOK, logs)
}
