package cmd

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"

	"github.com/lucasfabre/codegen/src/lua_runtime"
	"github.com/lucasfabre/codegen/src/processor"
	"github.com/spf13/cobra"
)

var cleanCmd = &cobra.Command{
	Use:   "clean [script.lua]",
	Short: "Empty the content of generated blocks",
	Long:  `Run a script in clean mode to empty all generated blocks and files it manages.`,
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		coordinator := processor.NewCoordinator(cfg)
		coordinator.CleanMode = true

		rt, err := luaruntime.New(cfg)
		if err != nil {
			return err
		}
		defer rt.Close()

		// Enable recursive processing
		rt.ProcessFunc = func(path, requestor string) error {
			return coordinator.Process(path, requestor)
		}
		rt.WaitFunc = coordinator.WaitForReader
		rt.ReadFunc = coordinator.GetResult

		// Case 1: File argument
		if len(args) > 0 {
			absEntry, _ := filepath.Abs(args[0])
			if filepath.Ext(absEntry) == ".lua" {
				if err := rt.ExecuteFile(absEntry); err != nil {
					return err
				}
				coordinator.Wait()
				rt.Schedule()
				if err := coordinator.CaptureResults(rt, absEntry); err != nil {
					return err
				}
			} else {
				if err := coordinator.Process(absEntry, ""); err != nil {
					return err
				}
				coordinator.Wait()
			}
			return coordinator.Commit()
		}

		// Case 2: Default entrypoint
		if entrypoint, ok := findDefaultEntrypoint(); ok {
			absEntry, _ := filepath.Abs(entrypoint)
			if err := rt.ExecuteFile(absEntry); err != nil {
				return err
			}
			coordinator.Wait()
			rt.Schedule()
			if err := coordinator.CaptureResults(rt, absEntry); err != nil {
				return err
			}
			return coordinator.Commit()
		}

		// Case 3: Stdin
		fi, _ := os.Stdin.Stat()
		if (fi.Mode() & os.ModeCharDevice) == 0 {
			scanner := bufio.NewScanner(os.Stdin)
			var script strings.Builder
			for scanner.Scan() {
				script.WriteString(scanner.Text())
				script.WriteString("\n")
			}
			if err := rt.DoString(script.String()); err != nil {
				return err
			}
			if err := coordinator.CaptureResults(rt, ""); err != nil {
				return err
			}
			coordinator.Wait()
			return coordinator.Commit()
		}

		return cmd.Help()
	},
}

func init() {
	rootCmd.AddCommand(cleanCmd)
}
