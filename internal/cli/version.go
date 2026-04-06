package cli

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number",
	RunE:  runVersion,
}

func runVersion(cmd *cobra.Command, args []string) error {
	if flagJSON {
		return printJSON(map[string]any{
			"version":   version,
			"os":        runtime.GOOS,
			"arch":      runtime.GOARCH,
			"commit":    commit,
			"buildDate": date,
		})
	}

	v := version
	if len(v) > 0 && v[0] != 'v' {
		v = "v" + v
	}
	fmt.Printf("tene %s (%s/%s)\n", v, runtime.GOOS, runtime.GOARCH)
	return nil
}
