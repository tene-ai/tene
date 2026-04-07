package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newSyncCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sync",
		Short: "Sync vault with Tene Cloud (push + pull)",
		Long: `Sync performs a pull followed by a push, ensuring your local and
remote vaults are in sync. Equivalent to running 'tene pull && tene push'.

Requires Pro plan and active login (tene login).`,
		RunE: runSync,
	}
	cmd.Flags().Bool("force", false, "Force sync, overwriting conflicts")
	cmd.Flags().String("api-url", envOrDefault("API_URL", "https://api.tene.sh"), "Tene Cloud API base URL")
	return cmd
}

func runSync(cmd *cobra.Command, args []string) error {
	// Check auth
	token, _ := loadAuthToken()
	if token == "" {
		fmt.Fprintln(cmd.ErrOrStderr(), `
  Tene Cloud Sync

  Push and pull your encrypted vault across devices.
  Zero-knowledge encryption ensures your secrets stay private.

  Get started:
    1. tene login     -- Sign in with GitHub
    2. tene push      -- Upload your vault
    3. tene pull      -- Download on another device

  Pro plan ($5/month) required for cloud sync.
  Learn more: https://tene.sh`)
		return nil
	}

	// Authenticated: run pull then push
	if !flagQuiet {
		fmt.Fprintln(cmd.ErrOrStderr(), "  Syncing...")
	}

	// Pull first
	if err := runPull(cmd, nil); err != nil {
		if !flagQuiet {
			fmt.Fprintf(cmd.ErrOrStderr(), "  Pull skipped: %v\n", err)
		}
	}

	// Then push
	if err := runPush(cmd, nil); err != nil {
		return err
	}

	if !flagQuiet {
		fmt.Fprintln(cmd.ErrOrStderr(), "  ✓ Sync complete")
	}
	return nil
}
