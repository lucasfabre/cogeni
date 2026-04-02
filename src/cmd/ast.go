package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/lucasfabre/cogeni/src/astparser"
	"github.com/spf13/cobra"
)

var astLang string

var astCmd = &cobra.Command{
	Use:   "ast <file>",
	Short: "Dump the AST of a file as JSON",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		filePath := args[0]

		lang := astLang
		if lang == "" {
			lang = cfg.GetGrammarForExtension(filepath.Ext(filePath))
		}

		source, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("failed to read file: %w", err)
		}

		parser, err := astparser.New(cfg.Grammar.Location)
		if err != nil {
			return fmt.Errorf("failed to initialize parser: %w", err)
		}

		ast, err := parser.Parse(lang, source)
		if err != nil {
			return fmt.Errorf("parsing error: %w", err)
		}

		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(ast)
	},
}

func init() {
	rootCmd.AddCommand(astCmd)
	astCmd.Flags().StringVarP(&astLang, "lang", "l", "", "Override language grammar (defaults to file extension)")
}
