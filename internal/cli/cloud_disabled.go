package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

const cloudDisabledMessage = `Cloud features are being redesigned.

tene works 100%% locally — no cloud needed:
  tene init      Create encrypted vault
  tene set KEY   Store a secret (interactive, hidden input)
  tene run --    Inject secrets into any command

Follow updates: https://tene.sh
`

// wrapCloudCmd replaces a command's RunE (and all subcommands recursively)
// with a function that prints a "coming soon" message and exits cleanly.
func wrapCloudCmd(cmd *cobra.Command) *cobra.Command {
	disableRunE := func(cmd *cobra.Command, args []string) error {
		fmt.Fprintf(os.Stderr, cloudDisabledMessage)
		return nil
	}

	var disableRecursive func(c *cobra.Command)
	disableRecursive = func(c *cobra.Command) {
		c.Run = nil
		c.RunE = disableRunE
		c.Args = nil // clear arg validation so message shows without args
		for _, sub := range c.Commands() {
			disableRecursive(sub)
		}
	}

	disableRecursive(cmd)
	return cmd
}
