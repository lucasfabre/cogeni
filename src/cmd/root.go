// Package cmd implements the command-line interface for cogeni,
// providing subcommands for AST parsing, Lua execution, and configuration management.
package cmd

import (
	"fmt"
	"os"

	"github.com/lucasfabre/codegen/src/config"
	"github.com/lucasfabre/codegen/src/processor"
	"github.com/spf13/cobra"
)

var (
	cfg *config.Config
)

var rootCmd = &cobra.Command{
	Use:     "cogeni",
	Short:   "cogeni is a language-agnostic code generation tool",
	Long:    `A powerful tool that regenerates code blocks using embedded Lua scripts and JQ queries.`,
	Version: "0.1.0",
	Args:    cobra.ArbitraryArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			// Search for default entrypoint in current directory
			if entrypoint, ok := findDefaultEntrypoint(); ok {
				ctx, err := newExecutionContext()
				if err != nil {
					return err
				}
				defer ctx.runtime.Close()

				if err := ctx.runEntrypoint(entrypoint); err != nil {
					return err
				}
				return ctx.coordinator.Commit()
			}
			return cmd.Help()
		}

		coordinator := processor.NewCoordinator(cfg)
		for _, arg := range args {
			if err := coordinator.Process(arg, ""); err != nil {
				return err
			}
		}

		coordinator.Wait()
		return coordinator.Commit()
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	var err error
	cfg, err = config.LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}
}
