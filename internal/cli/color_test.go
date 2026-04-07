package cli

import (
	"testing"
)

func TestColorEnabled_NoColorFlag(t *testing.T) {
	flagNoColor = true
	defer func() { flagNoColor = false }()

	if colorEnabled() {
		t.Error("colorEnabled() should return false when --no-color is set")
	}
}

func TestColorEnabled_NoColorEnv(t *testing.T) {
	flagNoColor = false
	t.Setenv("NO_COLOR", "1")

	if colorEnabled() {
		t.Error("colorEnabled() should return false when NO_COLOR env is set")
	}
}

func TestColorize_Disabled(t *testing.T) {
	flagNoColor = true
	defer func() { flagNoColor = false }()

	result := colorize(colorRed, "hello")
	if result != "hello" {
		t.Errorf("colorize() = %q, want %q (no color)", result, "hello")
	}
}

func TestColorize_AllFunctions_Disabled(t *testing.T) {
	flagNoColor = true
	defer func() { flagNoColor = false }()

	text := "test message"
	if got := redText(text); got != text {
		t.Errorf("redText() = %q, want %q", got, text)
	}
	if got := greenText(text); got != text {
		t.Errorf("greenText() = %q, want %q", got, text)
	}
	if got := yellowText(text); got != text {
		t.Errorf("yellowText() = %q, want %q", got, text)
	}
	if got := blueText(text); got != text {
		t.Errorf("blueText() = %q, want %q", got, text)
	}
	if got := boldText(text); got != text {
		t.Errorf("boldText() = %q, want %q", got, text)
	}
	if got := dimText(text); got != text {
		t.Errorf("dimText() = %q, want %q", got, text)
	}
}
