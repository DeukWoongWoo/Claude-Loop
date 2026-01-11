// Package main is the entry point for claude-loop CLI.
package main

import (
	"os"

	"github.com/DeukWoongWoo/claude-loop/internal/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		os.Exit(1)
	}
}
