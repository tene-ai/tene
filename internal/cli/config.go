package cli

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/agent-kay-it/tene/internal/vaultcfg"
)

// previewFrontConfirmText is the user-visible prompt shown when the user
// raises preview.front above zero. The wording is intentionally
// byte-identical to plan.md F1 step 7 + prd.md §5 Failure Mode G mitigation
// #3: the documents are the source of truth for security-sensitive copy,
// and a test asserts byte-equivalence so the prompt cannot drift.
const previewFrontConfirmText = "WARNING: setting preview.front > 0 will expose API key prefixes (sk-, ghp_, AKIA...) in vault.db.\n" +
	"This makes service identification possible if vault.db leaks. Continue? [y/N]"

// configKeyPrefix is the "config." namespace under which vault-scoped
// config keys live in vault_meta. Users type unprefixed keys
// ("preview.enabled") which we normalize via normalizeConfigKey before
// handing off to vaultcfg.
const configKeyPrefix = "config."

// normalizeConfigKey turns a user-friendly key into the canonical
// vaultcfg key. We accept both the bare form ("preview.enabled") and the
// fully-qualified form ("config.preview.enabled") so power users who
// have inspected vault_meta directly can use either spelling.
func normalizeConfigKey(userInput string) string {
	if strings.HasPrefix(userInput, configKeyPrefix) {
		return userInput
	}
	return configKeyPrefix + userInput
}

// denormalizeConfigKey strips the "config." prefix for user-visible
// display. We mirror what the user typed rather than echoing the
// internal namespace.
func denormalizeConfigKey(canonical string) string {
	return strings.TrimPrefix(canonical, configKeyPrefix)
}

var configFlagForce bool

var configCmd = &cobra.Command{
	Use:   "config [KEY[=VALUE]]",
	Short: "Show or modify vault-scoped configuration",
	Long: `Show or modify vault-scoped configuration keys stored in vault.db.

Without arguments, prints all known keys and their effective values.
With "KEY", prints only that key's value.
With "KEY=VALUE", validates and persists the new value.

Privacy:
  preview.enabled       Show partial value previews in 'tene list' (default: true)
  preview.front         Chars exposed from start (default: 0; max 8 total)
  preview.back          Chars exposed from end   (default: 4; max 8 total)

Audit:
  audit.warnAtMB        Warn when audit_log exceeds this size in MB (default: 50)

Setting preview.front > 0 requires explicit confirmation because it
exposes API key prefixes (sk-, ghp_, AKIA-) inside vault.db. Pass --force
to skip the prompt in scripted environments.`,
	Args: cobra.MaximumNArgs(1),
	RunE: runConfig,
}

func init() {
	configCmd.Flags().BoolVar(&configFlagForce, "force", false,
		"Skip the preview.front confirmation prompt")
}

func runConfig(cmd *cobra.Command, args []string) error {
	app, err := loadApp()
	if err != nil {
		return err
	}
	defer func() { _ = app.Vault.Close() }()

	// Zero-arg form: list everything.
	if len(args) == 0 {
		return printAllConfig(app)
	}

	arg := args[0]
	if !strings.Contains(arg, "=") {
		return printSingleConfig(app, arg)
	}

	// KEY=VALUE form. SplitN with N=2 lets values legitimately contain '='.
	parts := strings.SplitN(arg, "=", 2)
	rawKey := strings.TrimSpace(parts[0])
	value := strings.TrimSpace(parts[1])

	key := normalizeConfigKey(rawKey)
	if !vaultcfg.IsKnown(key) {
		return fmt.Errorf("unknown config key %q. Run `tene config` to list valid keys", rawKey)
	}

	// Confirm prompt for preview.front raising above zero. The check
	// happens BEFORE we ask vaultcfg to validate so we don't burn a
	// validation pass and so the prompt's "current value" message is
	// always accurate.
	if key == vaultcfg.KeyPreviewFront {
		if err := confirmPreviewFrontChange(app, value); err != nil {
			return err
		}
	}

	if err := vaultcfg.Set(app.Vault, key, value); err != nil {
		// vaultcfg returns sentinel errors that already carry user-readable
		// strings. Wrap minimally so the user sees the underlying message.
		if errors.Is(err, vaultcfg.ErrInvalidConfigValue) {
			return err
		}
		if errors.Is(err, vaultcfg.ErrInvalidConfigKey) {
			return err
		}
		return fmt.Errorf("failed to set %q: %w", key, err)
	}

	// Re-read so the success message reflects the canonical normalized
	// form (e.g. "yes" -> "true").
	stored, err := vaultcfg.Get(app.Vault, key)
	if err != nil {
		return err
	}

	if flagJSON {
		return printJSON(map[string]any{
			"ok":    true,
			"key":   key,
			"value": stored,
		})
	}
	if !flagQuiet {
		fmt.Printf("%s = %s\n", key, stored)
	}
	return nil
}

func printAllConfig(app *App) error {
	keys := vaultcfg.KnownKeys()
	values := make(map[string]string, len(keys))
	for _, k := range keys {
		val, err := vaultcfg.Get(app.Vault, k)
		if err != nil {
			return err
		}
		values[k] = val
	}

	if flagJSON {
		return printJSON(map[string]any{
			"ok":     true,
			"config": values,
		})
	}

	for _, k := range keys {
		fmt.Printf("%s = %s\n", k, values[k])
	}
	return nil
}

func printSingleConfig(app *App, key string) error {
	// FX6 B12: users naturally type the bare form (`tene config
	// preview.enabled`), but the stored key carries the `config.`
	// prefix. Try both shapes so the rc1 "unknown config key" error
	// only fires for genuinely unknown keys.
	resolved := key
	if !vaultcfg.IsKnown(resolved) {
		prefixed := "config." + key
		if vaultcfg.IsKnown(prefixed) {
			resolved = prefixed
		} else {
			return fmt.Errorf("unknown config key %q. Run `tene config` to list valid keys", key)
		}
	}
	val, err := vaultcfg.Get(app.Vault, resolved)
	if err != nil {
		return err
	}

	if flagJSON {
		return printJSON(map[string]any{
			"ok":    true,
			"key":   key,
			"value": val,
		})
	}
	fmt.Println(val)
	return nil
}

// confirmPreviewFrontChange enforces the security-sensitive interaction
// described in design.md §0 D6 and prd.md §5 Failure Mode G mitigation #3.
//
// We only prompt when:
//   - the requested value is a parseable integer > 0 (negative/non-numeric
//     are rejected later by vaultcfg.Set, so prompting them would confuse
//     the user);
//   - the current stored value is 0 (the safe default) OR strictly less
//     than the requested value (lowering or maintaining never prompts);
//   - --force was NOT passed.
//
// In non-interactive contexts (stdin is not a TTY) and without --force, we
// refuse rather than silently accept: a CI/CD that wants to set
// preview.front>0 must explicitly opt in to that risk via --force.
func confirmPreviewFrontChange(app *App, requested string) error {
	requestedN, err := strconv.Atoi(requested)
	if err != nil {
		// Let vaultcfg.Set produce the canonical "must be 0-8" error.
		return nil
	}
	if requestedN <= 0 {
		return nil
	}

	current, err := vaultcfg.CurrentPreviewFront(app.Vault)
	if err != nil {
		return err
	}
	if requestedN <= current {
		// Same or lower exposure level; no new risk to acknowledge.
		return nil
	}

	if configFlagForce {
		return nil
	}

	fmt.Fprintln(os.Stderr, previewFrontConfirmText)
	if !isTerminal() {
		return fmt.Errorf("preview.front > 0 requires interactive confirmation; pass --force to override")
	}

	reader := bufio.NewReader(os.Stdin)
	answer, _ := reader.ReadString('\n')
	answer = strings.TrimSpace(strings.ToLower(answer))
	if answer != "y" && answer != "yes" {
		return fmt.Errorf("aborted by user")
	}
	return nil
}
