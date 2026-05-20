// Package auth declares the permission tier model for the tene CLI.
//
// A permission tier captures the trust requirement of a single CLI verb,
// e.g. whether the command needs the user's Master Password (unlock) to
// fulfil its job. The tier of each command is declared exactly once in the
// CommandTier map below; the cobra dispatcher in internal/cli/root.go
// reads that map in its PersistentPreRunE hook and only triggers the
// unlock flow when the tier requires it.
//
// Why a separate package with a single static table?
//
//   - Single source of truth. The same table is exercised by unit tests
//     (every registered cobra command must have a tier — quality gate G4
//     in master-plan.md §5) and by the audit log (the tier name is part
//     of the cli.<tier>.<verb> action prefix recorded for each invocation
//     in F4).
//   - Panic-on-missing. Validate() walks the rootCmd subtree and refuses
//     to start the binary if any command is undeclared. New commands
//     therefore cannot ship without an explicit security review.
//   - Audit ergonomics. A reviewer can read this one file and know which
//     verbs decrypt plaintext (PermSecretRead), which write new
//     ciphertext (PermSecretWrite), and which touch only metadata
//     (PermMetaRead — list/env/audit/permissions/etc.).
//
// The 25 entries below mirror the CommandTier class diagram in
// docs/sprints/cli-ux-permission-model/design.md §1.1, with the
// `logout` entry removed as part of sprint v1014-rc1-qa-fixes FX4
// (B5: cloud verbs are intentionally not registered in root.go init
// while the cloud feature is being redesigned, so listing logout in
// the tier table made `tene permissions` advertise a verb the
// binary refuses to dispatch). The login/push/pull/sync/billing/team
// verbs were never in this table; when cloud is re-enabled, the PR
// that re-registers those verbs will also add their tier entries.
package auth

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/cobra"
)

// PermLevel classifies the trust requirement of a CLI command.
//
// The zero value is PermMetaRead deliberately: the safest default is "no
// unlock". Any command that needs the master key has to opt in by
// declaring a stronger tier in CommandTier.
type PermLevel int

const (
	// PermMetaRead reads only vault metadata — secret names, environment
	// names, schema version, the (already-plaintext) preview substring,
	// audit log rows. NEVER selects the encrypted_value column. No
	// password prompt, no keychain probe.
	//
	// Examples: list, env list, env create, permissions, audit tail.
	PermMetaRead PermLevel = iota

	// PermSecretWrite encrypts a new value into the vault. The encryption
	// subkey is HKDF-derived from the master key, so the master key is
	// required (keychain → env var → prompt fallback).
	//
	// Examples: set, import, init, delete, audit prune.
	PermSecretWrite

	// PermSecretRead decrypts an existing ciphertext back to plaintext.
	// STDOUT_SECRET_BLOCKED still applies on the call site (get/export/
	// run); this tier flag only says "the dispatcher must unlock".
	//
	// Examples: get, export, run, passwd, recover.
	PermSecretRead
)

// String returns a stable, lowercase token used in the audit log action
// prefix cli.<tier>.<verb> (see F4). Don't reorder — operators grep on it.
func (p PermLevel) String() string {
	switch p {
	case PermMetaRead:
		return "metaread"
	case PermSecretWrite:
		return "secretwrite"
	case PermSecretRead:
		return "secretread"
	default:
		return fmt.Sprintf("unknown(%d)", int(p))
	}
}

// RequiresUnlock reports whether a command of this tier needs the master
// key resolved (keychain → env var → prompt) before its RunE executes.
// Only PermMetaRead skips unlock; everything else requires it.
func (p PermLevel) RequiresUnlock() bool {
	return p != PermMetaRead
}

// CommandTier maps a cobra command path (space-joined, e.g. "env list",
// "audit prune") to its PermLevel.
//
// The path is what cobra's Command.CommandPath() yields minus the root
// "tene" segment. For top-level commands this is just the verb ("list",
// "set"); for sub-commands it's "<group> <verb>" ("env list", "audit
// tail"). The catch-all group commands themselves (e.g. "env" with no
// args, which falls through to env list) also get an entry so that
// Validate() accepts them.
//
// 25 entries total: 15 PermMetaRead + 5 PermSecretWrite + 5 PermSecretRead.
// Adding a new cobra command without adding an entry here causes Validate()
// to return an error which root.go's init() turns into a startup panic —
// see quality gate G4 (master-plan.md §5). Sprint v1014-rc1-qa-fixes/FX4
// extended Validate to also catch the reverse drift: a CommandTier entry
// whose path is not reachable in rootCmd is reported as a "stale entry"
// and refused at startup the same way.
var CommandTier = map[string]PermLevel{
	// --- PermMetaRead (15) -------------------------------------------------
	// No master-key unlock. Reads only metadata columns or static info.

	"list":        PermMetaRead, // F3 will rewrite to read preview column directly.
	"env":         PermMetaRead, // Catch-all root for env sub-tree; bare form switches env.
	"env list":    PermMetaRead,
	"env create":  PermMetaRead,
	"env delete":  PermMetaRead, // Deleting an env is a metadata-only op (cascade DELETE in SQL).
	"permissions": PermMetaRead, // F5 — table of all tiers; tier is reserved now so Validate() passes early.
	"whoami":      PermMetaRead,
	"version":     PermMetaRead,
	"update":      PermMetaRead,
	"completion":  PermMetaRead,
	"audit":       PermMetaRead, // F8 — catch-all root for audit sub-tree.
	"audit tail":  PermMetaRead, // F8.
	"audit show":  PermMetaRead, // F8.
	"config":      PermMetaRead, // F1 — preview.* / audit.* config keys; no value decryption.
	"migrate":     PermMetaRead, // F1 — schema status; fill-previews subcommand handles its own unlock.

	// --- PermSecretWrite (5) -----------------------------------------------
	// Master-key unlock required to encrypt a new plaintext into the vault.

	"set":         PermSecretWrite,
	"import":      PermSecretWrite,
	"delete":      PermSecretWrite, // Value never decrypted but row removal is a write op.
	"init":        PermSecretWrite, // Sets the master password; vault creation is a write.
	"audit prune": PermSecretWrite, // F8 — audit log row removal requires the write tier.

	// --- PermSecretRead (5) ------------------------------------------------
	// Master-key unlock required to decrypt plaintext. STDOUT_SECRET_BLOCKED
	// policy still applies at the call site (I-3, I-5).

	"get":     PermSecretRead,
	"export":  PermSecretRead,
	"run":     PermSecretRead,
	"passwd":  PermSecretRead, // Rotation needs to decrypt with the OLD key first.
	"recover": PermSecretRead, // Recovery flow re-derives + re-encrypts all secrets.
}

// TierFor returns the declared tier for a cobra command path. The boolean
// is false when the path is not in CommandTier — callers should treat
// that as a programmer error (see Validate()) and fail closed.
func TierFor(cmdPath string) (PermLevel, bool) {
	tier, ok := CommandTier[cmdPath]
	return tier, ok
}

// ErrMissingTier is wrapped by Validate() and surfaces in the panic
// message at binary startup, so the author of an unregistered new
// command sees a clear pointer to internal/auth/permissions.go.
var ErrMissingTier = errors.New("command has no PermLevel entry in internal/auth.CommandTier")

// Validate walks rootCmd and every reachable subcommand and confirms
// that each one has a CommandTier entry. This is the forward direction
// only: a registered verb without a tier declaration is reported. It is
// the check that init() has called since PR #116 to fail-loud at binary
// startup when a developer forgets to update CommandTier.
//
// Reverse-drift detection (a CommandTier entry whose verb is NOT
// registered in rootCmd — sprint v1014-rc1-qa-fixes/FX4, bug B5) lives
// in ValidateStrict so that synthetic-tree unit tests can exercise the
// forward path without having to populate every tier entry. Production
// init() calls ValidateStrict so both directions are enforced at the
// real binary's startup.
//
// The returned error lists every missing path on a separate line so
// the startup panic is actionable in one read.
func Validate(rootCmd *cobra.Command) error {
	if rootCmd == nil {
		return errors.New("Validate: rootCmd is nil")
	}

	var missing []string
	walk(rootCmd, "", &missing)

	if len(missing) == 0 {
		return nil
	}

	sort.Strings(missing)
	return fmt.Errorf(
		"%w: missing tier declaration(s):\n  - %s\nadd them to internal/auth.CommandTier",
		ErrMissingTier,
		strings.Join(missing, "\n  - "),
	)
}

// ValidateStrict runs Validate (forward direction) AND additionally
// asserts that every CommandTier entry is reachable in the rootCmd
// subtree. The reverse direction catches the class of bug that QA
// filed as B5 in v1.0.14-rc1: the `logout` entry was left in
// CommandTier after the cloud feature was unregistered from root.go's
// init(), so `tene permissions` advertised a verb the binary refused
// to dispatch.
//
// Forward and reverse drifts are reported with distinct prefixes
// ("missing tier declaration" vs "stale entry") so the operator knows
// the fix direction. The error wraps ErrMissingTier in both cases so
// callers (root.go init) can use errors.Is to recognise the class.
//
// Production startup uses ValidateStrict. Unit tests on synthetic
// cobra trees use the looser Validate to avoid having to materialise
// every CommandTier entry in their test fixtures.
func ValidateStrict(rootCmd *cobra.Command) error {
	if rootCmd == nil {
		return errors.New("ValidateStrict: rootCmd is nil")
	}

	var missing []string
	walk(rootCmd, "", &missing)
	stale := findStaleTierEntries(rootCmd)

	if len(missing) == 0 && len(stale) == 0 {
		return nil
	}

	sort.Strings(missing)
	sort.Strings(stale)

	var msg strings.Builder
	if len(missing) > 0 {
		msg.WriteString("missing tier declaration(s):\n  - ")
		msg.WriteString(strings.Join(missing, "\n  - "))
	}
	if len(stale) > 0 {
		if msg.Len() > 0 {
			msg.WriteString("\n")
		}
		msg.WriteString("stale entry(ies) in CommandTier (verb not registered in rootCmd):\n  - ")
		msg.WriteString(strings.Join(stale, "\n  - "))
	}
	msg.WriteString("\nedit internal/auth/permissions.go to fix")

	return fmt.Errorf("%w: %s", ErrMissingTier, msg.String())
}

// findStaleTierEntries returns the CommandTier paths that have no
// corresponding command in the rootCmd subtree. Used by Validate to
// detect the reverse-drift class of bug (B5).
//
// A path is considered "reachable" iff cobra.Find returns a command
// whose own Name() matches the last path segment AND whose CommandPath
// (minus the root name) equals the original tier path. The double
// check defends against cobra.Find returning the closest ancestor when
// the leaf is missing (e.g. for "env nonexistent" cobra returns envCmd
// rather than an error).
func findStaleTierEntries(rootCmd *cobra.Command) []string {
	var stale []string
	rootName := rootCmd.Name()
	for path := range CommandTier {
		parts := strings.Fields(path)
		if len(parts) == 0 {
			stale = append(stale, path)
			continue
		}
		cmd, _, err := rootCmd.Find(parts)
		if err != nil || cmd == nil || cmd == rootCmd {
			stale = append(stale, path)
			continue
		}
		// Walk back the actual command path (strip the root prefix) and
		// compare with the tier key. This catches the "cobra.Find returns
		// the ancestor instead of an error" case.
		actual := strings.TrimPrefix(cmd.CommandPath(), rootName+" ")
		if actual != path {
			stale = append(stale, path)
		}
	}
	return stale
}

// walk recurses through the cobra tree collecting every leaf or
// runnable group whose CommandTier entry is missing.
//
// We deliberately skip the synthetic "help" command (cobra adds it
// automatically; it never reaches PersistentPreRunE through user input
// in a way the tier table needs to police — `tene help foo` just
// formats the foo command's help text without running its RunE).
func walk(cmd *cobra.Command, parentPath string, missing *[]string) {
	for _, sub := range cmd.Commands() {
		// Compose the path as cobra would in CommandPath() minus the
		// root name. "tene env list" → "env list".
		path := sub.Name()
		if parentPath != "" {
			path = parentPath + " " + sub.Name()
		}

		// Skip cobra's auto-generated help command and any hidden
		// completion helpers like "completion bash" — the user-facing
		// "completion" verb itself IS in the table, but its dynamically
		// generated shell-specific children (bash/zsh/fish/powershell)
		// are cobra internals that share the parent's policy.
		if IsCobraSynthetic(sub) {
			continue
		}
		if path == "completion bash" || path == "completion zsh" ||
			path == "completion fish" || path == "completion powershell" {
			continue
		}

		if _, ok := CommandTier[path]; !ok {
			*missing = append(*missing, path)
		}

		// Recurse into groups. A command can be both a group (has
		// children) and a runnable verb (envCmd is the canonical
		// example: `tene env <name>` switches, `tene env list` lists).
		if sub.HasSubCommands() {
			walk(sub, path, missing)
		}
	}
}

// IsCobraSynthetic returns true for commands that cobra synthesizes
// (help, __complete, __completeNoDesc) and that should not be required
// to declare a tier — they never reach user-facing RunE through normal
// invocation.
//
// Exported because the runtime dispatcher in internal/cli/root.go uses
// the same predicate to short-circuit its tier lookup. Before sprint
// v1014-rc1-qa-fixes/FX4 the dispatcher tried to look "help" up in
// CommandTier and refused dispatch when it was not found, which is how
// `tene help` and `tene help set` ended up broken in v1.0.14-rc1 (B4).
// Sharing the predicate keeps the static and runtime checks in lockstep.
func IsCobraSynthetic(c *cobra.Command) bool {
	if c == nil {
		return false
	}
	switch c.Name() {
	case "help", "__complete", "__completeNoDesc":
		return true
	}
	return false
}
