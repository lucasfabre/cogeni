package cmd

import (
	"github.com/lucasfabre/cogeni/src/lua_runtime"
	"github.com/spf13/cobra"
)

var replCmd = &cobra.Command{
	Use:   "repl",
	Short: "Start an interactive Lua shell",
	RunE: func(cmd *cobra.Command, args []string) error {
		rt, err := luaruntime.New(cfg)
		if err != nil {
			return err
		}
		defer rt.Close()

		return rt.REPL()
	},
}

func init() {
	rootCmd.AddCommand(replCmd)
}
