package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tomo-kay/tene/internal/config"
)

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync vault to cloud (coming soon)",
	RunE:  runSync,
}

func runSync(cmd *cobra.Command, args []string) error {
	// Record sync attempt analytics
	_ = config.IncrementSyncAttempts()

	if flagJSON {
		return printJSON(map[string]any{
			"ok":          true,
			"message":     "cloud_sync_not_available",
			"waitlistUrl": "https://tene.sh/waitlist",
			"tip":         "Use tene export --encrypted for local backup.",
		})
	}

	fmt.Println(`
  Tene Cloud Sync -- Coming Soon!

  Solo ($5/month):
  - Cross-machine vault sync
  - Encrypted cloud backup (zero-knowledge)
  - Device management & audit log

  Team ($10/user/month):
  - Team secret sharing with RBAC
  - Environment-level permissions
  - Team dashboard & audit log

  Join the waitlist to get early access:
  -> https://tene.sh/waitlist

  In the meantime, use 'tene export --encrypted' for local backup.`)

	return nil
}
