package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"

	"github.com/spf13/cobra"
)

var diffCmd = &cobra.Command{
	Use:   "diff [file...]",
	Short: "Show changes that would be made",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, err := newExecutionContext()
		if err != nil {
			return err
		}
		defer ctx.runtime.Close()

		// 1. Run generation in memory
		if len(args) == 0 {
			if entrypoint, ok := findDefaultEntrypoint(); ok {
				if err := ctx.runEntrypoint(entrypoint); err != nil {
					return err
				}
			} else {
				return cmd.Help()
			}
		} else {
			for _, arg := range args {
				absArg, _ := filepath.Abs(arg)
				if filepath.Ext(absArg) == ".lua" {
					if err := ctx.runEntrypoint(absArg); err != nil {
						return err
					}
				} else {
					if err := ctx.coordinator.Process(absArg, ""); err != nil {
						return err
					}
				}
			}
			ctx.coordinator.Wait()
		}

		// 2. Diff results
		results := ctx.coordinator.GetResults()
		paths := make([]string, 0, len(results))
		for p := range results {
			paths = append(paths, p)
		}
		sort.Strings(paths)

		for _, path := range paths {
			newContent := results[path]
			if err := showDiff(path, newContent); err != nil {
				return err
			}
		}

		return nil
	},
}

func showDiff(path, newContent string) error {
	oldContent, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read %s: %w", path, err)
	}

	if string(oldContent) == newContent {
		return nil
	}

	// Create temp file for new content
	tmpFile, err := os.CreateTemp("", "cogeni-diff-*")
	if err != nil {
		return err
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(newContent); err != nil {
		return err
	}
	tmpFile.Close()

	// Use system diff for high quality colored output if available
	diffCmd := exec.Command("diff", "-u", "--color=always", path, tmpFile.Name())
	out, _ := diffCmd.CombinedOutput()
	fmt.Printf("--- %s\n+++ %s (generated)\n%s\n", path, path, string(out))

	return nil
}

func init() {
	rootCmd.AddCommand(diffCmd)
}
