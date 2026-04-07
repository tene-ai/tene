// Package response provides standard API response helpers.
package response

import (
	"errors"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/tomo-kay/tene/internal/domain"
)

type successResp struct {
	OK   bool `json:"ok"`
	Data any  `json:"data"`
	Meta meta `json:"meta"`
}

type errorResp struct {
	OK      bool   `json:"ok"`
	Error   string `json:"error"`
	Message string `json:"message"`
	Status  int    `json:"status"`
	Meta    meta   `json:"meta"`
}

type meta struct {
	Timestamp string `json:"timestamp"`
	RequestID string `json:"request_id,omitempty"`
}

func newMeta(c echo.Context) meta {
	return meta{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		RequestID: c.Response().Header().Get(echo.HeaderXRequestID),
	}
}

// OK sends a success response.
func OK(c echo.Context, status int, data any) error {
	return c.JSON(status, &successResp{
		OK:   true,
		Data: data,
		Meta: newMeta(c),
	})
}

// Err sends an error response, mapping domain errors to HTTP status codes.
func Err(c echo.Context, err error) error {
	status, code, msg := mapError(err)
	return c.JSON(status, &errorResp{
		OK:      false,
		Error:   code,
		Message: msg,
		Status:  status,
		Meta:    newMeta(c),
	})
}

// ErrMsg sends a custom error response.
func ErrMsg(c echo.Context, status int, code, message string) error {
	return c.JSON(status, &errorResp{
		OK:      false,
		Error:   code,
		Message: message,
		Status:  status,
		Meta:    newMeta(c),
	})
}

// M-02: Use generic messages to prevent user enumeration and info leakage.
func mapError(err error) (status int, code, msg string) {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		return http.StatusNotFound, "NOT_FOUND", "resource not found"
	case errors.Is(err, domain.ErrEmailAlreadyExists):
		return http.StatusConflict, "CONFLICT", "account conflict"
	case errors.Is(err, domain.ErrProjectAlreadyExists):
		return http.StatusConflict, "PROJECT_EXISTS", "project already exists"
	case errors.Is(err, domain.ErrUnauthorized),
		errors.Is(err, domain.ErrTokenExpired),
		errors.Is(err, domain.ErrTokenInvalid):
		return http.StatusUnauthorized, "UNAUTHORIZED", "authentication required"
	case errors.Is(err, domain.ErrForbidden):
		return http.StatusForbidden, "FORBIDDEN", "access denied"
	case errors.Is(err, domain.ErrProPlanRequired):
		return http.StatusPaymentRequired, "PRO_PLAN_REQUIRED", "pro plan required"
	case errors.Is(err, domain.ErrVaultNotFound):
		return http.StatusNotFound, "VAULT_NOT_FOUND", "vault not found"
	case errors.Is(err, domain.ErrVersionConflict):
		return http.StatusConflict, "VERSION_CONFLICT", "version conflict, pull first"
	case errors.Is(err, domain.ErrChecksumMismatch):
		return http.StatusBadRequest, "CHECKSUM_MISMATCH", "data integrity check failed"
	case errors.Is(err, domain.ErrNotTeamMember):
		return http.StatusForbidden, "NOT_TEAM_MEMBER", "not a team member"
	case errors.Is(err, domain.ErrInvalidOAuthCode),
		errors.Is(err, domain.ErrInvalidProvider):
		return http.StatusBadRequest, "BAD_REQUEST", "invalid request"
	default:
		return http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error"
	}
}
