// F5 — `tene permissions` info command.
//
// Renders the declarative permission tier map (internal/auth.CommandTier)
// as a human-readable table or a machine-readable JSON document. The
// command itself is registered as PermMetaRead in F2's tier map, so
// invoking it never prompts for the master password and never opens the
// vault for any decryption — it operates purely on the static table.
//
// Why a dedicated command rather than `tene --help`?
//
//   - `tene --help` is cobra's auto-generated synopsis; mutating it to
//     embed the tier table would couple help text formatting to the auth
//     package. Keeping permissions as its own verb lets `--json` produce
//     a structured shape that scripts and AI agents can parse.
//   - The tier table is the single source of truth for "which commands
//     need my password?". Surfacing it as a first-class CLI verb makes
//     that contract greppable by curious users and reviewable in CI
//     (TestPermissions_Text_AllEntriesPresent below asserts the verb's
//     output stays in sync with the table).
//
// Output format (text mode):
//
//	COMMAND                TIER           PASSWORD?
//	-----------------------------------------------
//	audit                  metaread              no
//	audit show             metaread              no
//	audit tail             metaread              no
//	completion             metaread              no
//	... (sorted alphabetically WITHIN each tier)
//	audit prune            secretwrite          yes
//	delete                 secretwrite          yes
//	... (PermSecretWrite block)
//	export                 secretread           yes
//	get                    secretread           yes
//	... (PermSecretRead block)
//
//	Total: 26 commands  (16 metaread / 5 secretwrite / 5 secretread)
//
// JSON mode (`--json`) shape:
//
//	{
//	  "ok": true,
//	  "count": 26,
//	  "byTier": { "metaread": 16, "secretwrite": 5, "secretread": 5 },
//	  "commands": [
//	    {"name": "audit",        "tier": "metaread",    "requiresUnlock": false},
//	    {"name": "audit prune",  "tier": "secretwrite", "requiresUnlock": true},
//	    ...
//	  ]
//	}
//
// The JSON `commands` array is sorted by (tier-order, name) so consumers
// can rely on a stable byte ordering. Tier order matches the text table:
// metaread → secretwrite → secretread.
package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"

	"github.com/spf13/cobra"

	"github.com/agent-kay-it/tene/internal/auth"
)

// permissionsCmd is the `tene permissions` top-level subcommand.
//
// Tier declaration: PermMetaRead (declared in internal/auth.CommandTier
// since F2; G4 validator already asserts this entry exists).
var permissionsCmd = &cobra.Command{
	Use:   "permissions",
	Short: "Show which commands require the master password",
	Long: `Show which commands require the master password.

tene classifies every CLI verb into one of three tiers:

  metaread     reads vault metadata only (names, environments, schema
               info, preview substring) — no password required.
  secretwrite  encrypts a new value into the vault — password required.
  secretread   decrypts an existing value — password required.

This command prints the full mapping. By design, AI assistants invoking
'tene list' or 'tene env list' never have to prompt the human for a
password, while 'tene get' and 'tene set' do.

Use --json for machine-readable output. The JSON 'commands' array is
sorted by (tier, name) so byte ordering is stable.

Permission tiers are hard-coded in internal/auth.CommandTier — they
cannot be overridden at runtime. See SECURITY.md for the rationale.`,
	RunE: runPermissions,
}

// tierOrder fixes the print order of the three tiers in BOTH text and
// JSON output. Operators grep on these tokens; reordering would break
// downstream pipelines that pin to "metaread" before "secretread".
var tierOrder = []auth.PermLevel{
	auth.PermMetaRead,
	auth.PermSecretWrite,
	auth.PermSecretRead,
}

// runPermissions is the RunE for `tene permissions`. It reads
// internal/auth.CommandTier exactly once, groups entries by tier in
// tierOrder, sorts alphabetically WITHIN each tier, and emits the
// requested representation.
//
// No I/O beyond stdout / stderr. No vault open. No keychain probe.
// (verified by TestPermissions_NoPasswordPrompt below.)
func runPermissions(cmd *cobra.Command, args []string) error {
	rows := collectPermissionRows()
	if flagJSON {
		return writePermissionsJSON(os.Stdout, rows)
	}
	return writePermissionsText(os.Stdout, rows)
}

// permissionRow is the in-memory shape of one CommandTier entry,
// flattened with the textual tier name for direct rendering.
type permissionRow struct {
	Name           string `json:"name"`
	Tier           string `json:"tier"`
	RequiresUnlock bool   `json:"requiresUnlock"`
	level          auth.PermLevel
}

// collectPermissionRows builds the sorted slice of permissionRow values
// from internal/auth.CommandTier. Sort order: tier first (using
// tierOrder positions), then alphabetical by command name.
//
// The function NEVER mutates auth.CommandTier (the map is read-only by
// contract — it is a package-level var literal seeded once at init).
func collectPermissionRows() []permissionRow {
	rows := make([]permissionRow, 0, len(auth.CommandTier))
	for name, level := range auth.CommandTier {
		rows = append(rows, permissionRow{
			Name:           name,
			Tier:           level.String(),
			RequiresUnlock: level.RequiresUnlock(),
			level:          level,
		})
	}

	tierIdx := make(map[auth.PermLevel]int, len(tierOrder))
	for i, t := range tierOrder {
		tierIdx[t] = i
	}

	sort.Slice(rows, func(i, j int) bool {
		if rows[i].level != rows[j].level {
			return tierIdx[rows[i].level] < tierIdx[rows[j].level]
		}
		return rows[i].Name < rows[j].Name
	})
	return rows
}

// writePermissionsText renders the rows as a fixed-width 3-column table.
// Column widths are computed from the row data so a future tier rename
// or a long sub-command path does not break alignment.
//
// Format choices:
//
//   - COMMAND column is left-aligned; the longest command path defines
//     the minimum width.
//   - TIER column is left-aligned; widest tier token defines the width.
//   - PASSWORD? column is right-aligned to keep the yes/no glyphs on
//     the right margin (easier visual scan).
//
// A footer line summarises the count totals per tier so a reader can
// sanity-check at a glance without piping into wc -l.
func writePermissionsText(w io.Writer, rows []permissionRow) error {
	cmdWidth := len("COMMAND")
	tierWidth := len("TIER")
	for _, r := range rows {
		if n := len(r.Name); n > cmdWidth {
			cmdWidth = n
		}
		if n := len(r.Tier); n > tierWidth {
			tierWidth = n
		}
	}

	// Header.
	if _, err := fmt.Fprintf(w, "%-*s  %-*s  %9s\n",
		cmdWidth, "COMMAND", tierWidth, "TIER", "PASSWORD?"); err != nil {
		return err
	}

	// Body.
	for _, r := range rows {
		pw := "no"
		if r.RequiresUnlock {
			pw = "yes"
		}
		if _, err := fmt.Fprintf(w, "%-*s  %-*s  %9s\n",
			cmdWidth, r.Name, tierWidth, r.Tier, pw); err != nil {
			return err
		}
	}

	// Footer count line. The "Total: N commands  (a metaread / b
	// secretwrite / c secretread)" format is asserted byte-for-byte
	// by TestPermissions_Text_Counts — keep it stable.
	counts := tierCounts(rows)
	if _, err := fmt.Fprintf(w, "\nTotal: %d commands  (%d metaread / %d secretwrite / %d secretread)\n",
		len(rows),
		counts[auth.PermMetaRead],
		counts[auth.PermSecretWrite],
		counts[auth.PermSecretRead],
	); err != nil {
		return err
	}
	return nil
}

// writePermissionsJSON emits the structured representation. We
// deliberately wrap the rows in a top-level object (rather than a bare
// array) so the shape mirrors the rest of the CLI's `--json` output
// surface (count + byTier + commands), which scripts can validate with
// a single JSON-Schema rule.
func writePermissionsJSON(w io.Writer, rows []permissionRow) error {
	counts := tierCounts(rows)
	// printJSON in root.go writes to os.Stdout directly; this helper
	// is parameterised on the writer to keep the test surface clean.
	doc := map[string]any{
		"ok":    true,
		"count": len(rows),
		"byTier": map[string]int{
			"metaread":    counts[auth.PermMetaRead],
			"secretwrite": counts[auth.PermSecretWrite],
			"secretread":  counts[auth.PermSecretRead],
		},
		"commands": rows,
	}
	if w == os.Stdout {
		return printJSON(doc)
	}
	// Test path: write with the same 2-space indent + trailing newline
	// shape as printJSON, but to a caller-supplied io.Writer.
	return encodeJSONIndented(w, doc)
}

// tierCounts returns a per-tier row count keyed by PermLevel.
func tierCounts(rows []permissionRow) map[auth.PermLevel]int {
	out := make(map[auth.PermLevel]int, len(tierOrder))
	for _, r := range rows {
		out[r.level]++
	}
	return out
}

// encodeJSONIndented mirrors printJSON's wire shape (2-space indent +
// trailing newline) but to a caller-supplied writer. Used by the unit
// tests so they can decode the output without juggling os.Stdout.
func encodeJSONIndented(w io.Writer, v any) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

// permissionsCmd is registered with rootCmd from internal/cli/root.go's
// init() so the verb is discoverable in one place alongside the rest of
// the command surface. The tier declaration ('permissions' = PermMetaRead)
// lives in internal/auth.CommandTier and was added in F2 — auth.Validate()
// asserts this entry exists at startup (G4 panic-on-missing).
