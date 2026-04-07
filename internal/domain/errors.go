// Package domain defines shared domain models and sentinel errors.
package domain

import "errors"

// Sentinel errors for domain logic. Handlers map these to HTTP status codes.
var (
	ErrNotFound           = errors.New("not found")
	ErrEmailAlreadyExists = errors.New("email already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUnauthorized       = errors.New("unauthorized")
	ErrForbidden          = errors.New("forbidden")
	ErrTokenExpired       = errors.New("token expired")
	ErrTokenInvalid       = errors.New("token invalid")
	ErrProPlanRequired    = errors.New("pro plan required")
	ErrInvalidOAuthCode   = errors.New("invalid oauth code")
	ErrInvalidProvider    = errors.New("invalid oauth provider")
	ErrVaultNotFound      = errors.New("vault not found")
	ErrVersionConflict    = errors.New("version conflict")
	ErrChecksumMismatch   = errors.New("checksum mismatch")
	ErrNotTeamMember      = errors.New("not a team member")
	ErrMergeConflict      = errors.New("merge conflict requires resolution")
	ErrProjectAlreadyExists = errors.New("project already exists")
)
