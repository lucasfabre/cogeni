package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var showCmd = &cobra.Command{
	Use:   "show",
	Short: "Visualize project information",
}

var showDepGraphCmd = &cobra.Command{
	Use:   "dependency-graph [script.lua]",
	Short: "Show the execution and dependency tree",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, err := newExecutionContext()
		if err != nil {
			return err
		}
		defer ctx.runtime.Close()

		entry := "cogeni.lua"
		if len(args) > 0 {
			entry = args[0]
		}

		// 1. Discovery Pass
		if _, err := os.Stat(entry); err == nil {
			if err := ctx.runEntrypoint(entry); err != nil {
				return err
			}

			// 2. Print Graph
			absEntry, _ := filepath.Abs(entry)
			printGraph(absEntry, ctx.coordinator.GetGraph())
		} else {
			return fmt.Errorf("script '%s' not found", entry)
		}

		return nil
	},
}

func printGraph(entry string, graph map[string][]string) {
	cwd, _ := os.Getwd()
	relEntry, err := filepath.Rel(cwd, entry)
	if err != nil {
		relEntry = entry
	}

	fmt.Println("Dependency Graph:")
	fmt.Printf("%s\n", relEntry)
	visited := make(map[string]bool)
	printNode(entry, graph, "  ", visited, cwd)
}

func printNode(path string, graph map[string][]string, indent string, visited map[string]bool, cwd string) {
	if visited[path] {
		return
	}
	visited[path] = true

	deps := graph[path]
	for i, dep := range deps {
		isLast := i == len(deps)-1
		prefix := "├── "
		nextIndent := indent + "│   "
		if isLast {
			prefix = "└── "
			nextIndent = indent + "    "
		}

		relDep, err := filepath.Rel(cwd, dep)
		if err != nil {
			relDep = dep
		}

		fmt.Printf("%s%s%s\n", indent, prefix, relDep)
		printNode(dep, graph, nextIndent, visited, cwd)
	}
}

func init() {
	rootCmd.AddCommand(showCmd)
	showCmd.AddCommand(showDepGraphCmd)
}
