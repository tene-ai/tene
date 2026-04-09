package errors

import (
	"encoding/json"
	"fmt"
	"io"
)

// TeneError is a structured CLI error.
// Code: machine-parseable error code (e.g., "VAULT_NOT_FOUND")
// Message: human-readable error message
// Exit: process exit code (0, 1, 2, 127)
type TeneError struct {
	Code    string `json:"error"`
	Message string `json:"message"`
	Exit    int    `json:"-"`
}

func (e *TeneError) Error() string {
	return e.Message
}

// WriteJSON outputs the error as JSON to the given writer.
func (e *TeneError) WriteJSON(w io.Writer) error {
	out := struct {
		OK      bool   `json:"ok"`
		Error   string `json:"error"`
		Message string `json:"message"`
	}{
		OK:      false,
		Error:   e.Code,
		Message: e.Message,
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(out)
}

// New creates a new TeneError.
func New(code, message string, exit int) *TeneError {
	return &TeneError{Code: code, Message: message, Exit: exit}
}

// Newf creates a new TeneError with a format string.
func Newf(code string, exit int, format string, args ...any) *TeneError {
	return &TeneError{
		Code:    code,
		Message: fmt.Sprintf(format, args...),
		Exit:    exit,
	}
}

// IsTeneError checks if an error is a *TeneError and returns it.
func IsTeneError(err error) (*TeneError, bool) {
	if te, ok := err.(*TeneError); ok {
		return te, true
	}
	return nil, false
}
