package main

import (
	"fmt"
	"os"
	"runtime/debug"

	"github.com/agent-kay-it/tene/internal/cli"
	teneerr "github.com/agent-kay-it/tene/pkg/errors"
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
		exitCode := 1 // default exit code

		if te, ok := teneerr.IsTeneError(err); ok {
			exitCode = te.Exit

			if hasJSONFlag() {
				_ = te.WriteJSON(os.Stderr)
			} else {
				fmt.Fprintf(os.Stderr, "Error: %s\n", te.Message)
			}
		} else {
			// Non-TeneError: plain error
			if hasJSONFlag() {
				fallback := teneerr.New("UNKNOWN_ERROR", err.Error(), 1)
				_ = fallback.WriteJSON(os.Stderr)
			} else {
				fmt.Fprintf(os.Stderr, "Error: %s\n", err)
			}
		}

		os.Exit(exitCode)
	}
}

// hasJSONFlag detects --json flag from os.Args before cobra parsing.
func hasJSONFlag() bool {
	for _, arg := range os.Args[1:] {
		if arg == "--json" {
			return true
		}
	}
	return false
}
