package errors

import (
	"bytes"
	"encoding/json"
	"fmt"
	"testing"
)

func TestTeneError_Error(t *testing.T) {
	err := ErrVaultNotFound
	want := "Not in a Tene project. Run \"tene init\" first."
	if got := err.Error(); got != want {
		t.Errorf("Error() = %q, want %q", got, want)
	}
}

func TestTeneError_WriteJSON(t *testing.T) {
	var buf bytes.Buffer
	err := ErrVaultNotFound
	if writeErr := err.WriteJSON(&buf); writeErr != nil {
		t.Fatal(writeErr)
	}

	var result map[string]any
	if jsonErr := json.Unmarshal(buf.Bytes(), &result); jsonErr != nil {
		t.Fatal(jsonErr)
	}

	if result["ok"] != false {
		t.Error("expected ok=false")
	}
	if result["error"] != "VAULT_NOT_FOUND" {
		t.Errorf("error = %v, want VAULT_NOT_FOUND", result["error"])
	}
	if result["message"] != ErrVaultNotFound.Message {
		t.Errorf("message = %v, want %q", result["message"], ErrVaultNotFound.Message)
	}
}

func TestIsTeneError(t *testing.T) {
	te, ok := IsTeneError(ErrVaultNotFound)
	if !ok || te.Code != "VAULT_NOT_FOUND" {
		t.Error("expected TeneError with VAULT_NOT_FOUND")
	}
}

func TestIsTeneError_PlainError(t *testing.T) {
	_, ok := IsTeneError(fmt.Errorf("some plain error"))
	if ok {
		t.Error("expected false for plain error")
	}
}

func TestErrSecretNotFound_Format(t *testing.T) {
	err := ErrSecretNotFound("API_KEY", "production")
	want := `Secret "API_KEY" not found in "production" environment.`
	if err.Error() != want {
		t.Errorf("got %q, want %q", err.Error(), want)
	}
	if err.Exit != 1 {
		t.Errorf("Exit = %d, want 1", err.Exit)
	}
}

func TestNewf_Format(t *testing.T) {
	err := Newf("TEST_CODE", 42, "value is %d and name is %q", 123, "test")
	if err.Code != "TEST_CODE" {
		t.Errorf("Code = %q, want TEST_CODE", err.Code)
	}
	if err.Exit != 42 {
		t.Errorf("Exit = %d, want 42", err.Exit)
	}
	want := `value is 123 and name is "test"`
	if err.Message != want {
		t.Errorf("Message = %q, want %q", err.Message, want)
	}
}

func TestErrorCodes_ExitValues(t *testing.T) {
	// Exit 0 errors
	exit0 := []*TeneError{ErrVaultAlreadyExists, ErrKeychainError}
	for _, e := range exit0 {
		if e.Exit != 0 {
			t.Errorf("%s.Exit = %d, want 0", e.Code, e.Exit)
		}
	}

	// Exit 1 errors
	exit1 := []*TeneError{
		ErrVaultNotFound, ErrInvalidEnvName, ErrEmptyValue, ErrValueTooLarge,
		ErrEncryptFailed, ErrPermissionDenied, ErrDiskFull, ErrInteractiveRequired,
		ErrInvalidBackupFile,
	}
	for _, e := range exit1 {
		if e.Exit != 1 {
			t.Errorf("%s.Exit = %d, want 1", e.Code, e.Exit)
		}
	}

	// Exit 1 function errors
	exit1Funcs := []*TeneError{
		ErrSecretNotFound("k", "e"), ErrSecretAlreadyExists("k"),
		ErrEnvironmentNotFound("e"), ErrEnvironmentAlreadyExists("e"),
		ErrEnvironmentProtected("e", "r"), ErrInvalidKeyName("k"),
		ErrReservedKeyName("k"), ErrFileNotFound("f"),
		ErrFileParse("f", 1, "d"),
	}
	for _, e := range exit1Funcs {
		if e.Exit != 1 {
			t.Errorf("%s.Exit = %d, want 1", e.Code, e.Exit)
		}
	}

	// Exit 2 errors
	exit2 := []*TeneError{
		ErrPasswordMismatch, ErrPasswordTooShort, ErrInvalidPassword,
		ErrInvalidRecoveryKey, ErrDecryptFailed,
	}
	for _, e := range exit2 {
		if e.Exit != 2 {
			t.Errorf("%s.Exit = %d, want 2", e.Code, e.Exit)
		}
	}

	// Exit 127
	cmd := ErrCommandNotFound("test")
	if cmd.Exit != 127 {
		t.Errorf("COMMAND_NOT_FOUND.Exit = %d, want 127", cmd.Exit)
	}
}
