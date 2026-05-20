package crypto

// MaxPreviewChars is the absolute upper bound on how many characters of a
// secret value may be exposed by a preview (front + back combined). This is
// a security invariant of the cli-ux-permission-model sprint (Q2 decision):
// no matter what the user configures, DerivePreview will refuse to produce a
// preview longer than this so an attacker who exfiltrates vault.db cannot
// reconstruct a realistic API key prefix + body.
const MaxPreviewChars = 8

// previewSafeFallback is the string returned whenever a preview cannot be
// safely derived from the input. It deliberately does not start with any
// known service prefix (sk-, ghp_, AKIA, ...) so it cannot be confused with
// a real partial value when a list output is read by humans or AI.
const previewSafeFallback = "*****"

// previewEllipsis is the visual separator between the front and back slices
// of the value. It is intentionally a single Unicode glyph so the output
// width is stable even when front == 0 (e.g. "...aBcD") or back == 0
// (e.g. "sk-p..."). It is NOT a literal ASCII "..." to make it grep-friendly
// in the codebase: search for it via "…" or the literal rune.
const previewEllipsis = "…"

// DerivePreview returns a short, visually identifiable substring of plaintext
// suitable for storage in vault.db's secrets.preview column.
//
// front and back specify how many runes from the start and end of plaintext
// are exposed. The combined cap is MaxPreviewChars (=8). Any one of these
// conditions yields the safe fallback string and no information about
// plaintext leaks:
//   - front < 0 or back < 0
//   - front + back > MaxPreviewChars
//   - len([]rune(plaintext)) < front + back + 1 (value is too short to mask
//     safely; e.g. a 6-rune value with front=2, back=4 would otherwise leak
//     the entire value with just a separator)
//
// When front == 0 and back == 0, DerivePreview returns the empty string.
// This is the documented "preview disabled" sentinel (Q2 always-string
// contract) and is distinct from the safe fallback "*****".
//
// The function is rune-aware (Unicode-safe): it slices by rune, never by
// byte, so multi-byte UTF-8 secrets do not produce mojibake. Caller is
// responsible for ensuring plaintext does not contain control characters
// it would not want echoed in a list output; DerivePreview does not strip
// or sanitize.
//
// The function never returns an error. Invalid inputs collapse to the safe
// fallback so call sites do not need to branch on configuration mistakes.
func DerivePreview(plaintext string, front, back int) string {
	// Defensive validation. These conditions are user-config driven (via
	// the `tene config preview.front=N` command), so callers may pass any
	// integer. We choose to return the safe fallback rather than panic so
	// a misconfiguration never makes `tene set` fail.
	if front < 0 || back < 0 {
		return previewSafeFallback
	}
	if front+back > MaxPreviewChars {
		return previewSafeFallback
	}
	if front == 0 && back == 0 {
		return ""
	}

	runes := []rune(plaintext)
	// Require at least one rune between front and back so the masked region
	// is non-empty. Without this, a 5-rune value with front=2, back=3 would
	// produce "ab" + ellipsis + "cde" which is the whole input.
	if len(runes) < front+back+1 {
		return previewSafeFallback
	}

	return string(runes[:front]) + previewEllipsis + string(runes[len(runes)-back:])
}
