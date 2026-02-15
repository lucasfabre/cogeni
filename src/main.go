// Package main is the entry point for the cogeni CLI tool.
// It initializes the command-line interface and delegates execution to the cmd package.
package main

import (
	"os"

	"github.com/lucasfabre/codegen/src/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
