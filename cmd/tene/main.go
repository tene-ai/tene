package main

import (
	"fmt"
	"os"
	"runtime/debug"

	"github.com/tomo-kay/tene/internal/cli"
)

// These are set via ldflags at build time (goreleaser).
var (
	version = ""
	commit  = ""
	date    = ""
)

func main() {
	// If ldflags didn't set version, try Go module build info (go install)
	if version == "" {
		if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" && info.Main.Version != "(devel)" {
			version = info.Main.Version
		} else {
			version = "dev"
		}
	}
	if commit == "" {
		if info, ok := debug.ReadBuildInfo(); ok {
			for _, s := range info.Settings {
				if s.Key == "vcs.revision" && len(s.Value) >= 7 {
					commit = s.Value[:7]
					break
				}
			}
		}
		if commit == "" {
			commit = "unknown"
		}
	}
	if date == "" {
		if info, ok := debug.ReadBuildInfo(); ok {
			for _, s := range info.Settings {
				if s.Key == "vcs.time" {
					date = s.Value
					break
				}
			}
		}
		if date == "" {
			date = "unknown"
		}
	}

	cli.SetVersion(version, commit, date)

	if err := cli.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}
}
