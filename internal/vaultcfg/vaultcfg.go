// Package vaultcfg manages vault-scoped configuration keys stored inside
// vault.db's vault_meta table under the "config." prefix.
//
// Why a separate package (not internal/config):
//
//   - internal/config owns the global ~/.tene/config.json file (preferences
//     that follow the user across projects). vaultcfg owns per-vault
//     preferences that follow the vault.db wherever it goes.
//
//   - Privacy controls like preview.enabled, preview.front, preview.back
//     describe how much of a secret value is exposed to readers of THIS
//     vault.db. Storing them in vault_meta means a vault.db copied to
//     another machine retains the privacy choice; a global JSON would
//     not.
//
//   - audit.warnAtMB describes the size threshold of THIS vault's
//     audit_log table. Co-locating the threshold with the data it
//     watches keeps the data flow obvious.
//
// All keys live under the "config." namespace in vault_meta to avoid
// collisions with reserved keys like "schema_version", "kdf_salt", etc.
package vaultcfg

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/agent-kay-it/tene/internal/vault"
)

// Keys are the fully-qualified vault_meta keys (with the "config." prefix
// already applied). Exported so the CLI layer can reference them by name
// without spelling the prefix repeatedly.
const (
	KeyPreviewEnabled = "config.preview.enabled"
	KeyPreviewFront   = "config.preview.front"
	KeyPreviewBack    = "config.preview.back"
	KeyAuditWarnAtMB  = "config.audit.warnAtMB"
)

// Defaults are the values used when a key has never been explicitly set.
// They are also the values seeded into Settings by GetPreviewSettings /
// GetAuditWarnAtMB so callers get a complete struct on a fresh vault.
//
// DefaultPreviewFront=0 is the security-conscious default chosen during
// the cli-ux-permission-model sprint (Q2 decision, 2026-05-20): a fresh
// vault never exposes API key prefixes (sk-, ghp_, AKIA-) so a leaked
// vault.db does not allow service identification. Users who want the
// visual cue of a prefix must explicitly run `tene config preview.front=N`
// (which triggers the confirm prompt) to opt in.
const (
	DefaultPreviewEnabled = true
	DefaultPreviewFront   = 0
	DefaultPreviewBack    = 4
	DefaultAuditWarnAtMB  = 50
)

// MaxPreviewFront / MaxPreviewBack bound a single config field. The sum
// (front+back) is bounded by pkg/crypto.MaxPreviewChars and validated in
// validatePreviewFront / validatePreviewBack via crossCheck.
const (
	maxPreviewFront    = 8
	maxPreviewBack     = 8
	maxPreviewCombined = 8 // mirrors pkg/crypto.MaxPreviewChars
	minAuditWarnAtMB   = 1
	maxAuditWarnAtMB   = 1000
)

// ErrInvalidConfigKey is returned when a key is not in the known set.
// Returning a sentinel here (rather than free-form fmt.Errorf) lets the CLI
// distinguish "user typo" from "vault read failure" and produce a tailored
// error message.
var ErrInvalidConfigKey = errors.New("vaultcfg: unknown config key")

// ErrInvalidConfigValue is returned when a value fails per-key validation
// (e.g. preview.front=99, audit.warnAtMB=0). It wraps the underlying reason
// so callers can either format the chained string for the user or
// errors.Is for tests.
var ErrInvalidConfigValue = errors.New("vaultcfg: invalid config value")

// PreviewSettings is the resolved triple used by `tene set` / `tene import`
// to drive pkg/crypto.DerivePreview. Returned as a value type (no pointers)
// so callers do not have to nil-check after the load helper.
type PreviewSettings struct {
	Enabled bool
	Front   int
	Back    int
}

// IsKnown reports whether key is a recognized config key. CLI layer uses
// this to fail fast on typos before any vault write.
func IsKnown(key string) bool {
	switch key {
	case KeyPreviewEnabled, KeyPreviewFront, KeyPreviewBack, KeyAuditWarnAtMB:
		return true
	default:
		return false
	}
}

// KnownKeys returns the full set of known config keys in display order.
// Stable order matters so `tene config` produces deterministic output.
func KnownKeys() []string {
	return []string{
		KeyPreviewEnabled,
		KeyPreviewFront,
		KeyPreviewBack,
		KeyAuditWarnAtMB,
	}
}

// Get returns the raw string value stored for key. If the key has not been
// set explicitly, the zero-value string-encoded default is returned (e.g.
// "true" for preview.enabled). This way callers never have to branch on
// "is this missing" vs "is this set" -- they always see a value.
func Get(v *vault.Vault, key string) (string, error) {
	if !IsKnown(key) {
		return "", fmt.Errorf("%w: %q", ErrInvalidConfigKey, key)
	}
	val, err := v.GetMeta(key)
	if err == nil {
		return val, nil
	}
	if errors.Is(err, vault.ErrMetaNotFound) {
		return defaultStringFor(key), nil
	}
	return "", err
}

// Set validates value and writes it to vault_meta under key.
//
// The validation happens BEFORE the write so a malformed user input never
// produces a half-applied config (we cannot rollback an UPSERT after the
// fact). Per-key normalization (bool aliases like "yes"/"no") happens here
// as well, so the stored form is always canonical ("true"/"false").
//
// Returns ErrInvalidConfigKey or a wrapped ErrInvalidConfigValue when the
// user input is bad; otherwise propagates the vault write error.
func Set(v *vault.Vault, key, value string) error {
	if !IsKnown(key) {
		return fmt.Errorf("%w: %q", ErrInvalidConfigKey, key)
	}

	normalized, err := validateAndNormalize(v, key, value)
	if err != nil {
		return err
	}

	if err := v.SetMeta(key, normalized); err != nil {
		return fmt.Errorf("vaultcfg: write %q: %w", key, err)
	}
	return nil
}

// GetPreviewSettings returns all three preview controls in one call. Used
// by `tene set` / `tene import` on every write so they can derive the
// preview from plaintext before encryption.
//
// Returns sensible defaults on a vault that has never seen `tene config`,
// so the helper is safe to call unconditionally.
func GetPreviewSettings(v *vault.Vault) (PreviewSettings, error) {
	enabled, err := getBool(v, KeyPreviewEnabled, DefaultPreviewEnabled)
	if err != nil {
		return PreviewSettings{}, err
	}
	front, err := getInt(v, KeyPreviewFront, DefaultPreviewFront)
	if err != nil {
		return PreviewSettings{}, err
	}
	back, err := getInt(v, KeyPreviewBack, DefaultPreviewBack)
	if err != nil {
		return PreviewSettings{}, err
	}
	// Defense-in-depth: if a corrupt vault somehow stored a value past the
	// hard cap, collapse it to the safe defaults here rather than passing
	// the bad pair down to DerivePreview (which would fall back to "*****"
	// for every secret silently). This keeps `tene list` looking sensible
	// while the user investigates the corruption.
	if front < 0 || back < 0 || front+back > maxPreviewCombined {
		front = DefaultPreviewFront
		back = DefaultPreviewBack
	}
	return PreviewSettings{Enabled: enabled, Front: front, Back: back}, nil
}

// GetAuditWarnAtMB returns the audit_log size threshold (in MB) at which
// the CLI should print a one-line stderr hint. Falls back to the default
// when unset or malformed.
func GetAuditWarnAtMB(v *vault.Vault) int {
	n, err := getInt(v, KeyAuditWarnAtMB, DefaultAuditWarnAtMB)
	if err != nil || n < minAuditWarnAtMB || n > maxAuditWarnAtMB {
		return DefaultAuditWarnAtMB
	}
	return n
}

// --- internal helpers ---

func defaultStringFor(key string) string {
	switch key {
	case KeyPreviewEnabled:
		return strconv.FormatBool(DefaultPreviewEnabled)
	case KeyPreviewFront:
		return strconv.Itoa(DefaultPreviewFront)
	case KeyPreviewBack:
		return strconv.Itoa(DefaultPreviewBack)
	case KeyAuditWarnAtMB:
		return strconv.Itoa(DefaultAuditWarnAtMB)
	default:
		return ""
	}
}

func validateAndNormalize(v *vault.Vault, key, value string) (string, error) {
	switch key {
	case KeyPreviewEnabled:
		b, err := parseBool(value)
		if err != nil {
			return "", fmt.Errorf("%w: %s must be true/false, got %q",
				ErrInvalidConfigValue, key, value)
		}
		return strconv.FormatBool(b), nil

	case KeyPreviewFront:
		n, err := strconv.Atoi(value)
		if err != nil {
			return "", fmt.Errorf("%w: %s must be 0-%d, got %q",
				ErrInvalidConfigValue, key, maxPreviewFront, value)
		}
		if n < 0 || n > maxPreviewFront {
			return "", fmt.Errorf("%w: %s must be 0-%d, got %d",
				ErrInvalidConfigValue, key, maxPreviewFront, n)
		}
		// Cross-check against the current preview.back so the combined
		// total never exceeds the hard cap. We read the OTHER axis from
		// the vault rather than requiring callers to pass it -- this keeps
		// validation centralized.
		other, err := getInt(v, KeyPreviewBack, DefaultPreviewBack)
		if err != nil {
			return "", err
		}
		if n+other > maxPreviewCombined {
			return "", fmt.Errorf("%w: preview.front (%d) + preview.back (%d) must not exceed %d",
				ErrInvalidConfigValue, n, other, maxPreviewCombined)
		}
		return strconv.Itoa(n), nil

	case KeyPreviewBack:
		n, err := strconv.Atoi(value)
		if err != nil {
			return "", fmt.Errorf("%w: %s must be 0-%d, got %q",
				ErrInvalidConfigValue, key, maxPreviewBack, value)
		}
		if n < 0 || n > maxPreviewBack {
			return "", fmt.Errorf("%w: %s must be 0-%d, got %d",
				ErrInvalidConfigValue, key, maxPreviewBack, n)
		}
		other, err := getInt(v, KeyPreviewFront, DefaultPreviewFront)
		if err != nil {
			return "", err
		}
		if n+other > maxPreviewCombined {
			return "", fmt.Errorf("%w: preview.front (%d) + preview.back (%d) must not exceed %d",
				ErrInvalidConfigValue, other, n, maxPreviewCombined)
		}
		return strconv.Itoa(n), nil

	case KeyAuditWarnAtMB:
		n, err := strconv.Atoi(value)
		if err != nil {
			return "", fmt.Errorf("%w: %s must be %d-%d, got %q",
				ErrInvalidConfigValue, key, minAuditWarnAtMB, maxAuditWarnAtMB, value)
		}
		if n < minAuditWarnAtMB || n > maxAuditWarnAtMB {
			return "", fmt.Errorf("%w: %s must be %d-%d, got %d",
				ErrInvalidConfigValue, key, minAuditWarnAtMB, maxAuditWarnAtMB, n)
		}
		return strconv.Itoa(n), nil
	}
	// Unreachable: IsKnown(key) was checked before we got here.
	return "", fmt.Errorf("%w: %q", ErrInvalidConfigKey, key)
}

// parseBool accepts the canonical boolean spellings plus the common shell
// aliases ("yes", "no", "on", "off"). strconv.ParseBool already handles
// "1"/"0"/"t"/"f"/"true"/"false"; we extend it for friendliness.
func parseBool(s string) (bool, error) {
	switch s {
	case "yes", "Yes", "YES", "on", "On", "ON":
		return true, nil
	case "no", "No", "NO", "off", "Off", "OFF":
		return false, nil
	default:
		return strconv.ParseBool(s)
	}
}

func getBool(v *vault.Vault, key string, fallback bool) (bool, error) {
	s, err := v.GetMeta(key)
	if err != nil {
		if errors.Is(err, vault.ErrMetaNotFound) {
			return fallback, nil
		}
		return false, err
	}
	b, err := parseBool(s)
	if err != nil {
		// Stored value is corrupt; fall back rather than failing the
		// command. This matches the resilience choice in GetPreviewSettings.
		return fallback, nil
	}
	return b, nil
}

func getInt(v *vault.Vault, key string, fallback int) (int, error) {
	s, err := v.GetMeta(key)
	if err != nil {
		if errors.Is(err, vault.ErrMetaNotFound) {
			return fallback, nil
		}
		return 0, err
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return fallback, nil
	}
	return n, nil
}

// CurrentPreviewFront reads preview.front directly (no defense-in-depth
// reset like GetPreviewSettings does). Used by the `tene config` confirm
// prompt to compare the OLD value vs the requested NEW one and decide
// whether to require user confirmation.
func CurrentPreviewFront(v *vault.Vault) (int, error) {
	return getInt(v, KeyPreviewFront, DefaultPreviewFront)
}
