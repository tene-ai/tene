//go:build generate
// +build generate

// Standalone tool used by GoReleaser's `before:` hook to write tene man
// pages into the `manpages/` directory for archive inclusion and Homebrew
// formula installation.
//
// Run locally with: go run ./scripts/gen-manpage.go

package main

import (
	"log"
	"os"

	"github.com/spf13/cobra/doc"
	tenecli "github.com/tene-ai/tene/internal/cli"
)

func main() {
	root := tenecli.RootCmd()

	header := &doc.GenManHeader{
		Title:   "TENE",
		Section: "1",
		Source:  "tene CLI",
		Manual:  "Tene Manual",
	}

	dir := "manpages"
	if err := os.MkdirAll(dir, 0o755); err != nil {
		log.Fatal(err)
	}
	if err := doc.GenManTree(root, header, dir); err != nil {
		log.Fatal(err)
	}
	log.Printf("Wrote man pages to %s/", dir)
}
