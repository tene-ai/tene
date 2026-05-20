package crypto

import "testing"

func TestDerivePreview_DefaultFrontZeroBackFour(t *testing.T) {
	// Q2 decision (2026-05-20): default front=0, back=4 means an API key
	// prefix like "sk-proj-" is never exposed; only the last 4 chars are.
	cases := []struct {
		name  string
		input string
		want  string
	}{
		{"openai-style", "sk-proj-1234567890aBcD", "…aBcD"},
		{"github-style", "ghp_abc123def456ghi789", "…i789"},
		{"aws-style", "AKIAIOSFODNN7EXAMPLE", "…MPLE"},
		{"exact-min-len-5", "hello", "…ello"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := DerivePreview(tc.input, 0, 4)
			if got != tc.want {
				t.Fatalf("DerivePreview(%q, 0, 4) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}

func TestDerivePreview_OptInFrontFourBackFour(t *testing.T) {
	// User explicitly opts into prefix exposure with `tene config preview.front=4`
	// (after the warning prompt). Combined 8 chars is at the hard cap.
	got := DerivePreview("sk-proj-1234567890aBcD", 4, 4)
	want := "sk-p…aBcD"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestDerivePreview_HardCapExceeded(t *testing.T) {
	// front + back must be <= MaxPreviewChars (8). Anything larger collapses
	// to the safe fallback. This is invariant I-9 (master-plan §8).
	cases := []struct {
		front, back int
	}{
		{9, 0},
		{0, 9},
		{5, 5}, // sum=10
		{4, 5}, // sum=9
		{8, 1}, // sum=9
		{99, 99},
	}
	for _, tc := range cases {
		got := DerivePreview("sk-proj-1234567890aBcD", tc.front, tc.back)
		if got != previewSafeFallback {
			t.Errorf("front=%d back=%d: got %q, want %q (hard cap)", tc.front, tc.back, got, previewSafeFallback)
		}
	}
}

func TestDerivePreview_NegativeInputs(t *testing.T) {
	// Negative front/back should never leak any character.
	cases := []struct {
		front, back int
	}{
		{-1, 0},
		{0, -1},
		{-1, -1},
		{-99, 4},
		{4, -99},
	}
	for _, tc := range cases {
		got := DerivePreview("sk-proj-1234567890aBcD", tc.front, tc.back)
		if got != previewSafeFallback {
			t.Errorf("front=%d back=%d: got %q, want %q (negatives)", tc.front, tc.back, got, previewSafeFallback)
		}
	}
}

func TestDerivePreview_BothZeroIsEmptyString(t *testing.T) {
	// front=0 and back=0 is the "preview disabled" sentinel, not the safe
	// fallback. JSON consumers expect an empty string and the always-string
	// contract (Q2) treats it distinctly from "*****".
	got := DerivePreview("anything-here", 0, 0)
	if got != "" {
		t.Fatalf("DerivePreview(..., 0, 0) = %q, want empty string", got)
	}
}

func TestDerivePreview_ValueTooShortToMask(t *testing.T) {
	// We require len(runes) >= front+back+1 so the masked middle is
	// guaranteed to be at least one rune wide. Otherwise we would be
	// echoing the whole value with just an ellipsis stuck in.
	cases := []struct {
		input       string
		front, back int
	}{
		{"", 0, 4},      // empty value
		{"abc", 0, 4},   // 3 runes, need 5
		{"abcd", 0, 4},  // 4 runes, need 5
		{"abcde", 2, 4}, // 5 runes, need 7
		{"a", 0, 4},
	}
	for _, tc := range cases {
		got := DerivePreview(tc.input, tc.front, tc.back)
		if got != previewSafeFallback {
			t.Errorf("input=%q front=%d back=%d: got %q, want %q",
				tc.input, tc.front, tc.back, got, previewSafeFallback)
		}
	}
}

func TestDerivePreview_UnicodeRuneSafe(t *testing.T) {
	// Multi-byte UTF-8: ensure we slice by rune, not by byte, so we never
	// split a code point. The Korean letters used here are 3 bytes each in
	// UTF-8.
	got := DerivePreview("자가가가가호호호호하", 0, 4)
	want := "…호호호하"
	if got != want {
		t.Fatalf("unicode rune slice: got %q, want %q", got, want)
	}

	got = DerivePreview("ab가나다라마", 2, 2)
	want = "ab…라마"
	if got != want {
		t.Fatalf("unicode rune slice mixed: got %q, want %q", got, want)
	}
}

func TestDerivePreview_FrontOnly(t *testing.T) {
	// back=0 with front>0 should emit "<front-chars>...". Tests the symmetry
	// of the helper.
	got := DerivePreview("sk-proj-1234567890aBcD", 4, 0)
	want := "sk-p…"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestDerivePreview_BackOnly(t *testing.T) {
	// front=0 with back>0 is the default config and the most common path.
	got := DerivePreview("0123456789", 0, 3)
	want := "…789"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestDerivePreview_BoundaryAtMaxCap(t *testing.T) {
	// front+back == MaxPreviewChars (=8) is allowed; one more is not.
	if got := DerivePreview("0123456789ABCDEF", 4, 4); got != "0123…CDEF" {
		t.Fatalf("at-cap front=4 back=4: got %q", got)
	}
	if got := DerivePreview("0123456789ABCDEF", 8, 0); got != "01234567…" {
		t.Fatalf("at-cap front=8 back=0: got %q", got)
	}
	if got := DerivePreview("0123456789ABCDEF", 0, 8); got != "…89ABCDEF" {
		t.Fatalf("at-cap front=0 back=8: got %q", got)
	}
}

func TestDerivePreview_NeverLeaksFullValue(t *testing.T) {
	// Invariant: the output (when not empty and not fallback) must always
	// contain the ellipsis separator. This guards against future refactors
	// that might accidentally return the whole input.
	previews := []string{
		DerivePreview("sk-proj-abcdef", 0, 4),
		DerivePreview("sk-proj-abcdef", 4, 4),
		DerivePreview("sk-proj-abcdef", 4, 0),
	}
	for _, p := range previews {
		if p == "" || p == previewSafeFallback {
			continue
		}
		if !containsRune(p, '…') {
			t.Errorf("preview %q is missing the ellipsis separator", p)
		}
	}
}

func containsRune(s string, want rune) bool {
	for _, r := range s {
		if r == want {
			return true
		}
	}
	return false
}
