// F8 — `tene audit` subcommand surface.
//
// Three subcommands sit under the `audit` parent:
//
//   - `tene audit tail [-n N] [--json]`   — last N rows, newest first.
//   - `tene audit show [--since DUR] [--filter PAT] [--json]` — filtered query.
//   - `tene audit prune --older-than DUR [--force] [--dry-run]` — delete old rows.
//
// Tier mapping (declared once in internal/auth.CommandTier and asserted
// at startup by F2's auth.Validate):
//
//   - audit             PermMetaRead   (catch-all root, prints help)
//   - audit tail        PermMetaRead   (read-only metadata query)
//   - audit show        PermMetaRead   (same — filtered query)
//   - audit prune       PermSecretWrite (destructive — requires master-key unlock)
//
// `audit prune`'s PermSecretWrite tier is declared centrally but F2
// deliberately leaves master-key unlock to each subcommand's RunE.
// runAuditPrune therefore calls loadOrPromptMasterKey before executing
// the DELETE so the contract is honoured at the place users feel it
// (the password prompt). The actual SQL deletion lives in
// vault.PruneAuditLog (the G10 chokepoint).
//
// NDJSON output choice (`--json`): one JSON object per line, no
// surrounding array, no commas between rows. This matches the
// append-only stream nature of audit_log and is friendly to `jq -c`,
// `grep`, and incremental consumers that read line by line. Wrapping
// in a JSON array would force callers to load the whole result into
// memory before parsing.
package cli

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/agent-kay-it/tene/internal/audit"
	"github.com/agent-kay-it/tene/pkg/crypto"
)

// Flags scoped to the audit subcommands. Module-level so resetFlags()
// in the test harness can zero them between runs (mirrors how F2 +
// F4 handled their command-specific flags).
var (
	auditTailN          int
	auditShowSince      string
	auditShowFilter     string
	auditShowResource   string
	auditShowLimit      int
	auditPruneOlderThan string
	auditPruneForce     bool
	auditPruneDryRun    bool
)

// auditCmd is the catch-all `tene audit` parent. With no subcommand
// it prints help. Tier is PermMetaRead — invoking the parent itself
// touches no vault data beyond what cobra's help renderer does.
var auditCmd = &cobra.Command{
	Use:   "audit",
	Short: "Inspect and prune the local audit log",
	Long: `Inspect and prune the audit_log table inside vault.db.

The audit log records every CLI invocation (action + permission tier)
plus per-verb domain events (secret.read, secret.write, vault.init,
etc.). It exists for forensic recall — answers to "what was attempted?"
and "when did the key holder access this secret?" — and is never
auto-deleted (master-plan §10 invariant I-14).

Subcommands:
  tail   Show the most recent N audit entries (default: 20).
  show   Filtered query by time window and action pattern.
  prune  Manually delete entries older than a given age (requires
         master-password unlock and explicit confirm / --force).

All audit commands except 'prune' run at the metadata-read tier —
they do NOT require the master password. 'prune' is destructive and
requires the password.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Falling through with no subcommand is identical to cobra's
		// default help rendering. We do not write our own row here —
		// the F4 dispatcher already recorded cli.metaread.audit when
		// PersistentPreRunE fired.
		return cmd.Help()
	},
}

var auditTailCmd = &cobra.Command{
	Use:   "tail",
	Short: "Show the most recent N audit entries (default: 20)",
	Long: `Show the most recent N audit entries, newest first.

The default of 20 rows is enough to see what the previous command
did and the F4 dispatcher row that recorded it. Use a larger -n
(e.g. -n 200) when investigating a specific session.

In --json mode, output is NDJSON (one JSON object per line, no
surrounding array). This is jq-friendly and matches the append-only
shape of audit_log.`,
	RunE: runAuditTail,
}

var auditShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Filtered audit-log query (by time window + action pattern + resource)",
	Long: `Filtered audit-log query.

Flags:
  --since DURATION    Match rows whose timestamp is within the last
                      DURATION. Examples: 1h, 24h, 7d (parsed via
                      Go time.ParseDuration; "d" is converted to 24h).
  --filter PATTERN    Match action column against a SQL-LIKE pattern.
                      Use '%' as wildcard. Examples:
                        --filter 'cli.metaread.%'
                        --filter 'cli.%.set'
                        --filter 'secret.write'
  --resource NAME     Substring match on the resource_name column.
                      Internally wrapped as %NAME%. Combine with
                      --filter for AND semantics. Examples:
                        --resource STRIPE_KEY
                        --resource STRIPE --filter 'cli.secretread.%'
  --limit N           Cap the result to N rows (default: 200).

In --json mode, output is NDJSON (see 'tene audit tail').`,
	RunE: runAuditShow,
}

var auditPruneCmd = &cobra.Command{
	Use:   "prune",
	Short: "Delete audit entries older than --older-than (destructive)",
	Long: `Delete audit entries older than --older-than.

The audit log grows roughly proportional to CLI activity. The 50 MB
threshold warning (--quiet to suppress) hints when prune is worth
running. There is no automatic rotation; this command is the only way
to remove rows (master-plan §10 invariant I-14).

Flags:
  --older-than DURATION  Required. Match rows whose timestamp is
                         strictly older than now - DURATION.
                         Examples: 30d, 90d, 1h (testing).
  --force                Skip the interactive confirmation prompt.
  --dry-run              Report how many rows would be deleted and
                         exit. No DELETE is issued.

Safety:
  - Requires master-password unlock (PermSecretWrite tier).
  - Without --force you are prompted to confirm. Pressing anything
    other than 'y' or 'yes' aborts with 0 rows deleted.
  - --dry-run never deletes; pairing with --force is meaningless.

Examples:
  tene audit prune --older-than 30d
  tene audit prune --older-than 90d --force
  tene audit prune --older-than 30d --dry-run`,
	RunE: runAuditPrune,
}

func init() {
	// Bind subcommands to parent before registering with rootCmd so
	// cobra's CommandPath() returns "tene audit tail" for the audit
	// emission + tier lookup.
	auditCmd.AddCommand(auditTailCmd)
	auditCmd.AddCommand(auditShowCmd)
	auditCmd.AddCommand(auditPruneCmd)

	auditTailCmd.Flags().IntVarP(&auditTailN, "n", "n", 20, "Number of rows to show")

	auditShowCmd.Flags().StringVar(&auditShowSince, "since", "", "Match rows within the last DURATION (e.g. 1h, 24h, 7d)")
	auditShowCmd.Flags().StringVar(&auditShowFilter, "filter", "", "SQL-LIKE pattern for the action column (e.g. 'cli.metaread.%')")
	auditShowCmd.Flags().StringVar(&auditShowResource, "resource", "", "Substring match on the resource_name column (LIKE %NAME%)")
	auditShowCmd.Flags().IntVar(&auditShowLimit, "limit", 200, "Maximum rows to return")

	auditPruneCmd.Flags().StringVar(&auditPruneOlderThan, "older-than", "", "Required. Delete rows older than DURATION (e.g. 30d, 90d)")
	auditPruneCmd.Flags().BoolVar(&auditPruneForce, "force", false, "Skip interactive confirmation")
	auditPruneCmd.Flags().BoolVar(&auditPruneDryRun, "dry-run", false, "Report row count without deleting")

	rootCmd.AddCommand(auditCmd)
}

// runAuditTail implements `tene audit tail`.
func runAuditTail(cmd *cobra.Command, args []string) error {
	app, err := loadApp()
	if err != nil {
		return err
	}
	defer func() { _ = app.Vault.Close() }()

	if auditTailN <= 0 {
		return fmt.Errorf("-n must be a positive integer (got %d)", auditTailN)
	}

	mgr := audit.New(app.Vault)
	rows, err := mgr.Tail(auditTailN)
	if err != nil {
		return err
	}

	if flagJSON {
		return writeNDJSON(os.Stdout, rows)
	}
	return writeAuditText(os.Stdout, rows)
}

// runAuditShow implements `tene audit show`.
func runAuditShow(cmd *cobra.Command, args []string) error {
	app, err := loadApp()
	if err != nil {
		return err
	}
	defer func() { _ = app.Vault.Close() }()

	f := audit.Filter{Limit: auditShowLimit}
	if auditShowSince != "" {
		d, err := parseHumanDuration(auditShowSince)
		if err != nil {
			return fmt.Errorf("--since %q: %w", auditShowSince, err)
		}
		f.Since = time.Now().UTC().Add(-d)
	}
	if auditShowFilter != "" {
		f.ActionMatch = auditShowFilter
	}
	// --resource is a user-friendly substring match; wrap as
	// %NAME% here so the vault layer's SQL LIKE call matches any
	// resource_name containing the substring. design.md §6B.1 +
	// plan.md F8 step 3 spec.
	if auditShowResource != "" {
		f.Resource = "%" + auditShowResource + "%"
	}

	mgr := audit.New(app.Vault)
	rows, err := mgr.Show(f)
	if err != nil {
		return err
	}

	if flagJSON {
		return writeNDJSON(os.Stdout, rows)
	}
	return writeAuditText(os.Stdout, rows)
}

// runAuditPrune implements `tene audit prune`.
//
// Flow:
//
//  1. Validate --older-than (required, positive).
//  2. Count candidate rows (cheap COUNT — no DELETE yet).
//  3. If --dry-run: print count, exit. Master-key unlock is NOT
//     required for dry-run since no destructive op happens.
//  4. Honour PermSecretWrite contract: prompt for master password
//     (via loadOrPromptMasterKey). The key is not actually used —
//     audit_log is plaintext SQL — but proving knowledge of the
//     master password is the user-facing gate for destructive ops.
//  5. If not --force: interactive confirm. Anything other than y/yes
//     aborts.
//  6. Call mgr.Prune(); report rows-deleted count.
//  7. Reset the threshold-warning sentinel so the next size check
//     fires again if the pruned vault is still over threshold.
func runAuditPrune(cmd *cobra.Command, args []string) error {
	if auditPruneOlderThan == "" {
		return fmt.Errorf("--older-than is required (e.g. --older-than 30d)")
	}
	d, err := parseHumanDuration(auditPruneOlderThan)
	if err != nil {
		return fmt.Errorf("--older-than %q: %w", auditPruneOlderThan, err)
	}
	if d <= 0 {
		return fmt.Errorf("--older-than must be a positive duration (got %v)", d)
	}

	app, err := loadApp()
	if err != nil {
		return err
	}
	defer func() { _ = app.Vault.Close() }()

	mgr := audit.New(app.Vault)
	count, err := mgr.CountOlderThan(d)
	if err != nil {
		return err
	}

	if !flagQuiet && !flagJSON {
		cutoff := time.Now().UTC().Add(-d).Format("2006-01-02 15:04:05")
		fmt.Fprintf(os.Stderr,
			"About to delete %d audit log row(s) older than %s UTC.\n",
			count, cutoff)
	}

	if auditPruneDryRun {
		if flagJSON {
			return printJSON(map[string]any{
				"ok":      true,
				"dryRun":  true,
				"matched": count,
			})
		}
		if !flagQuiet {
			fmt.Println("Dry run: no rows deleted.")
		}
		return nil
	}

	if count == 0 {
		if flagJSON {
			return printJSON(map[string]any{
				"ok":      true,
				"deleted": 0,
			})
		}
		if !flagQuiet {
			fmt.Println("Nothing to prune.")
		}
		return nil
	}

	// PermSecretWrite contract: prove knowledge of the master
	// password before any destructive audit_log mutation. We do not
	// USE the key (audit_log has no encrypted columns) — the prompt
	// is the gate.
	masterKey, err := loadOrPromptMasterKey(app)
	if err != nil {
		return err
	}
	defer crypto.ZeroBytes(masterKey)

	if !auditPruneForce {
		if !confirmDestructive(os.Stdin, os.Stderr, "Proceed?") {
			if !flagQuiet && !flagJSON {
				fmt.Fprintln(os.Stderr, "Aborted. No rows deleted.")
			}
			if flagJSON {
				return printJSON(map[string]any{
					"ok":       true,
					"deleted":  0,
					"aborted":  true,
				})
			}
			return nil
		}
	}

	deleted, err := mgr.Prune(d)
	if err != nil {
		return err
	}

	// Reset the threshold sentinel so the size check can fire again
	// if the user is still over the configured limit. resetAuditSentinel
	// silently ignores errors — the sentinel is purely an idempotency
	// hint, not a security boundary.
	resetAuditSentinel(app.Dir)

	if flagJSON {
		return printJSON(map[string]any{
			"ok":      true,
			"deleted": deleted,
		})
	}
	if !flagQuiet {
		fmt.Printf("Deleted %d row(s).\n", deleted)
	}
	return nil
}

// writeNDJSON renders rows in NDJSON form. Each row is a JSON object on
// its own line; we deliberately do NOT wrap in a JSON array because
// the audit_log is conceptually an append-only stream and NDJSON
// matches that consumer pattern.
//
// The on-wire timestamp is RFC 3339 UTC ("2026-05-20T14:23:01Z"),
// which is what consumers like jq + log-shipping pipelines expect.
func writeNDJSON(w io.Writer, rows []audit.LogEntry) error {
	bw := bufio.NewWriter(w)
	defer func() { _ = bw.Flush() }()

	enc := json.NewEncoder(bw)
	for _, r := range rows {
		// Re-shape to the wire form so json.Encoder emits the
		// timestamp field (LogEntry.Timestamp has json:"-").
		wire := struct {
			ID        int64  `json:"id"`
			Timestamp string `json:"ts"`
			Action    string `json:"action"`
			Resource  string `json:"resource"`
			Details   string `json:"details"`
		}{
			ID:        r.ID,
			Timestamp: r.Timestamp.UTC().Format(time.RFC3339),
			Action:    r.Action,
			Resource:  r.Resource,
			Details:   r.Details,
		}
		if err := enc.Encode(wire); err != nil {
			return err
		}
	}
	return nil
}

// writeAuditText renders rows in the human text format. Format:
//
//	2026-05-20 14:23:01  cli.metaread.list           default
//	2026-05-20 14:23:15  cli.secretwrite.set         STRIPE_KEY
//
// Columns: timestamp (UTC, second granularity) · action · resource.
// Details is omitted in text mode because no current audit row uses
// it; if F-future adds structured details we will format them here.
func writeAuditText(w io.Writer, rows []audit.LogEntry) error {
	bw := bufio.NewWriter(w)
	defer func() { _ = bw.Flush() }()

	if len(rows) == 0 {
		// Print a single hint so a caller piping the output never
		// confuses "empty result" with "command silently failed".
		_, err := fmt.Fprintln(bw, "(no matching audit entries)")
		return err
	}
	for _, r := range rows {
		ts := r.Timestamp.UTC().Format("2006-01-02 15:04:05")
		if _, err := fmt.Fprintf(bw, "%s  %-30s  %s\n", ts, r.Action, r.Resource); err != nil {
			return err
		}
	}
	return nil
}

// parseHumanDuration extends time.ParseDuration to accept a "d" suffix
// for days. Plain time.ParseDuration tops out at hours, which forces
// users to type "720h" for "30d". We translate "d" to "24h" before
// delegating so "30d" → 720h works.
//
// Combined forms like "1d12h" are accepted because the translation
// only rewrites the trailing 'd' segments; the suffix arithmetic is
// linear (one pass). For sprint scope we only need the simple "30d"
// case but tests against the parser cover edge inputs.
func parseHumanDuration(s string) (time.Duration, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, fmt.Errorf("empty duration")
	}
	// Quick path: stdlib accepts the value as-is.
	if d, err := time.ParseDuration(s); err == nil {
		return d, nil
	}
	// Slow path: split on 'd' segments and convert each to hours.
	// "30d"   -> "720h"
	// "1d12h" -> "24h12h" -> stdlib sums them.
	var out strings.Builder
	num := strings.Builder{}
	for _, r := range s {
		switch {
		case r >= '0' && r <= '9', r == '.':
			num.WriteRune(r)
		case r == 'd':
			n := num.String()
			num.Reset()
			if n == "" {
				return 0, fmt.Errorf("malformed duration %q (lonely 'd')", s)
			}
			// 24 * value hours. ParseFloat would let us handle "1.5d"
			// but we stick to integer days for clarity — fractional
			// days are odd in audit windows.
			out.WriteString(n)
			out.WriteString("d_BAD_") // sentinel: stdlib will reject if we reach here
			// Actually convert: replace last sentinel via direct math.
			// Simpler: emit "<n>*24h" — but ParseDuration won't do
			// multiplication. So we do the math ourselves below.
			//
			// To keep this readable, fall back to: parse n as int,
			// emit fmt.Sprintf("%dh", n*24).
			//
			// Discard the sentinel and reset.
			cur := out.String()
			cur = strings.TrimSuffix(cur, n+"d_BAD_")
			out.Reset()
			out.WriteString(cur)
			// Re-parse n as a numeric value to multiply.
			var days int64
			for _, c := range n {
				if c == '.' {
					return 0, fmt.Errorf("fractional days not supported in %q", s)
				}
				days = days*10 + int64(c-'0')
			}
			out.WriteString(fmt.Sprintf("%dh", days*24))
		default:
			// Other unit characters (h, m, s, ms, us, ns, µ) — write
			// them along with any pending numeric prefix.
			if num.Len() > 0 {
				out.WriteString(num.String())
				num.Reset()
			}
			out.WriteRune(r)
		}
	}
	if num.Len() > 0 {
		// Trailing digits with no unit — invalid.
		return 0, fmt.Errorf("malformed duration %q (trailing digits without unit)", s)
	}
	return time.ParseDuration(out.String())
}

// confirmDestructive prompts the user on stderr and reads y/N from
// stdin. Returns true iff the user typed "y" or "yes" (case
// insensitive). EOF / read errors are treated as "no" — defensive
// default for a destructive op.
//
// The prompt is on stderr (not stdout) so an automation pipeline that
// captures stdout still surfaces the question to the human operator
// before the program blocks on stdin.
func confirmDestructive(in io.Reader, errOut io.Writer, prompt string) bool {
	_, _ = fmt.Fprintf(errOut, "%s [y/N]: ", prompt)
	br := bufio.NewReader(in)
	line, err := br.ReadString('\n')
	if err != nil && err != io.EOF {
		return false
	}
	line = strings.ToLower(strings.TrimSpace(line))
	return line == "y" || line == "yes"
}
