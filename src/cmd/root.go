// Package cmd implements the command-line interface for cogeni,
// providing subcommands for AST parsing, Lua execution, and configuration management.
package cmd

import (
	"fmt"
	"os"

	"github.com/lucasfabre/cogeni/src/config"
	"github.com/lucasfabre/cogeni/src/processor"
	"github.com/spf13/cobra"
)

var (
	cfg  *config.Config
	jobs int
)

var rootCmd = &cobra.Command{
	Use:     "cogeni",
	Short:   "Generate reproducible artifacts from code and specs",
	Long:    `A programmable runtime for generating and synchronizing derived artifacts from source code and machine-readable specs.`,
	Version: "0.1.0",
	Args:    cobra.ArbitraryArgs,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if jobs > 0 && cfg != nil {
			cfg.Concurrency = jobs
		}
		return nil
	},
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
	rootCmd.PersistentFlags().IntVarP(&jobs, "jobs", "j", 0, "Number of concurrent jobs/runtimes (default 10 or configured)")
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
