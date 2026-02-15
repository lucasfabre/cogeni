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

var runEval string

var runCmd = &cobra.Command{
	Use:   "run [script.lua]",
	Short: "Execute a Lua script",
	Long:  `Run a Lua script file, execute inline code with -e, or read from stdin.`,
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		coordinator := processor.NewCoordinator(cfg)
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

		// Case 1: Eval flag
		if runEval != "" {
			if err := rt.DoString(runEval); err != nil {
				return err
			}
			coordinator.Wait()
			rt.Schedule()
			if err := coordinator.CaptureResults(rt, ""); err != nil {
				return err
			}
			return coordinator.Commit()
		}

		// Case 2: File argument
		if len(args) > 0 {
			absEntry, _ := filepath.Abs(args[0])
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
			coordinator.Wait()
			rt.Schedule()
			if err := coordinator.CaptureResults(rt, ""); err != nil {
				return err
			}
			return coordinator.Commit()
		}

		return cmd.Help()
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
	runCmd.Flags().StringVarP(&runEval, "eval", "e", "", "Execute inline Lua code")
}
