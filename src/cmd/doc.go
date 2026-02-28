package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

var docCmd = &cobra.Command{
	Use:    "doc [output_dir]",
	Short:  "Generate CLI documentation",
	Hidden: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		outputDir := "docs/docusaurus/docs/cli"
		if len(args) > 0 {
			outputDir = args[0]
		}

		err := os.MkdirAll(outputDir, 0755)
		if err != nil {
			return err
		}

		// Write an index file
		filename := outputDir + "/_category_.json"
		f, err := os.Create(filename)
		if err != nil {
			return err
		}
		defer f.Close()

		f.WriteString(`{
  "label": "CLI Reference",
  "position": 2,
  "link": {
    "type": "generated-index",
    "slug": "cli-reference",
    "description": "CLI Reference for cogeni commands."
  }
}`)

		err = doc.GenMarkdownTree(rootCmd, outputDir)
		if err != nil {
			return err
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(docCmd)
}
