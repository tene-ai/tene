package cli

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/agent-kay-it/tene/internal/keychain"
)

// F6 (Keychain Fallback UX Polish) — emit a one-time stderr notice when
// the CLI falls back to the file-based keystore because the OS keychain
// is unavailable. See docs/sprints/cli-ux-permission-model/design.md §4.3
// and plan.md §F6.
//
// Layering:
//   - Detection (was the file fallback used?) lives in package keychain
//     (returned via FallbackInfo from NewStoreWithStatus).
//   - Policy (when to emit, sentinel I/O, --quiet handling) lives here in
//     the CLI package — keychain stays free of cobra / cli concerns.

// sentinelDirName / sentinelPrefix together produce
//
//	~/.tene/.fallback-warned-<dir-hash>
//
// where dir-hash is HashPath(projectDir)[:12]. Per-project isolation
// matters: a user who has 3 different vaults on 3 different projects
// (all on the same machine with no keychain) gets exactly one notice
// per project, not one notice total. That matches the design's
// "one-time per machine + per vault project" goal.
const (
	sentinelDirName = ".tene"
	sentinelPrefix  = ".fallback-warned-"
)

// fallbackSentinelPath returns the absolute path of the sentinel file
// that marks "we have already warned the user about file-keystore
// fallback for this project". Returns an error only when the user's
// home directory cannot be resolved — which on a real install is
// catastrophic enough that the caller should just skip the notice
// rather than report (we cannot write the notice or the sentinel
// without HOME anyway).
func fallbackSentinelPath(projectDir string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return "", fmt.Errorf("resolve home dir: %w", err)
	}
	hash := keychain.HashPath(projectDir)
	return filepath.Join(home, sentinelDirName, sentinelPrefix+hash), nil
}

// emitFallbackNoticeIfNeeded prints a one-time stderr notice when:
//   - info.Used is true (the file-keystore fallback is in effect), AND
//   - the sentinel for this project does not already exist, AND
//   - --quiet is not set.
//
// On the happy path (keychain available), this is a no-op.
//
// Errors during sentinel I/O are silently swallowed: this function MUST
// NOT fail any command. A keychain notice is purely informational.
//
// stderr (not stdout) is mandatory so JSON consumers parsing stdout
// remain unaffected. Tests assert this separation.
//
// The --quiet branch deliberately does NOT write the sentinel — that
// way the next non-quiet run still gets the notice once.
//
// Concurrency: sentinel creation uses O_CREATE|O_EXCL so two tene
// processes racing to emit the notice both attempt the write, but only
// one wins. The loser sees "file exists" from OpenFile, treats it as
// "already warned", and skips the print. This is the desired behavior
// — the user sees the notice at most once even under parallel exec.
func emitFallbackNoticeIfNeeded(stderr io.Writer, info keychain.FallbackInfo, projectDir string, quiet bool) {
	if !info.Used {
		return
	}
	if quiet {
		return
	}

	sentinel, err := fallbackSentinelPath(projectDir)
	if err != nil {
		// Cannot resolve HOME -> cannot write sentinel -> would notify on
		// every run. That is bad UX but better than crashing the CLI.
		// Print once per process and move on.
		writeFallbackNotice(stderr, info.Path)
		return
	}

	// If sentinel already exists, the user has been notified before for
	// this project. Stay silent.
	if _, statErr := os.Stat(sentinel); statErr == nil {
		return
	}

	// Race-safe sentinel creation. O_EXCL guarantees we only "win" once.
	if err := os.MkdirAll(filepath.Dir(sentinel), 0o700); err != nil {
		// Cannot create ~/.tene/. Print the notice but skip the sentinel
		// — we will reprint next time, which is annoying but harmless.
		writeFallbackNotice(stderr, info.Path)
		return
	}
	f, err := os.OpenFile(sentinel, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	if err != nil {
		// Lost the race (another tene process just created it) OR a
		// genuine filesystem error. Either way: do not double-print.
		// If lost-race: the winning process is printing the notice
		// concurrently. If filesystem error: best-effort, don't spam.
		return
	}
	_ = f.Close()

	writeFallbackNotice(stderr, info.Path)
}

// writeFallbackNotice produces the actual user-visible text. Kept as a
// separate function so tests can hit the same byte sequence without
// reinventing the format string.
//
// Format choices:
//   - "note:" lowercase prefix matches Go stdlib conventions for
//     informational stderr lines (cf. "go: ..." messages).
//   - Trailing period inside the parens, then a hint sentence outside.
//     Reads naturally when piped to a log file.
func writeFallbackNotice(stderr io.Writer, keyfilePath string) {
	_, _ = fmt.Fprintf(stderr,
		"note: tene is using file-based keystore at %s (keychain unavailable).\n",
		keyfilePath)
	_, _ = fmt.Fprintln(stderr,
		"      This message shows once per project. Pass --quiet to suppress.")
}
