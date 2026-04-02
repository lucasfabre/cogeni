package tests

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/lucasfabre/cogeni/src/astparser"
)

var testGrammarDir string

func TestMain(m *testing.M) {
	// Use a persistent directory for grammars to avoid re-downloading/re-compiling
	// across test runs and test cases.
	// We use .cache/test-grammars in the project root (assuming tests are run from project root or tests dir)

	// Determine project root. Since this test is in "tests/", parent is root.
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	projectRoot := filepath.Dir(wd)
	if filepath.Base(wd) != "tests" {
		// If running from root via go test ./...
		projectRoot = wd
	}

	testGrammarDir = filepath.Join(projectRoot, ".cache", "test-grammars")

	if err := os.MkdirAll(testGrammarDir, 0755); err != nil {
		panic(err)
	}

	code := m.Run()

	// Optional: Clean up if desired, but keeping it speeds up future runs
	// os.RemoveAll(testGrammarDir)

	os.Exit(code)
}

func TestParseTypeScript(t *testing.T) {
	parser, err := astparser.New(testGrammarDir)
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	// Source file to parse
	sourcePath := filepath.Join("..", "examples", "fullstack", "frontend", "api.ts")
	source, err := os.ReadFile(sourcePath)
	if err != nil {
		t.Fatalf("Failed to read test file %s: %v", sourcePath, err)
	}

	// Parse the file
	rootNode, err := parser.Parse("typescript", source)
	if err != nil {
		t.Fatalf("Failed to parse typescript: %v", err)
	}

	// Basic validation
	if rootNode.Type != "program" {
		t.Errorf("Expected root node type 'program', got '%s'", rootNode.Type)
	}

	if len(rootNode.Children) == 0 {
		t.Error("Expected children in root node, got 0")
	}

	// Look for the class declaration (it's inside an export_statement field)
	foundClass := false
	for _, child := range rootNode.Children {
		if child.Type == "export_statement" {
			if decl, ok := child.Fields["declaration"]; ok {
				if node, ok := decl.(astparser.Node); ok && node.Type == "class_declaration" {
					foundClass = true
					break
				}
			}
		}
	}

	if !foundClass {
		t.Error("Expected to find a 'class_declaration' (inside 'export_statement' field) in the AST")
	}
}

func TestParseGo(t *testing.T) {
	parser, err := astparser.New(testGrammarDir)
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	// Sample Go source
	source := []byte(`package main

func main() {}
`)

	// Parse the source
	rootNode, err := parser.Parse("go", source)
	if err != nil {
		t.Fatalf("Failed to parse go: %v", err)
	}

	// Basic validation
	if rootNode.Type != "source_file" {
		t.Errorf("Expected root node type 'source_file', got '%s'", rootNode.Type)
	}

	foundPackage := false
	for _, child := range rootNode.Children {
		if child.Type == "package_clause" {
			foundPackage = true
			break
		}
	}

	if !foundPackage {
		t.Error("Expected to find a 'package_clause' in the AST")
	}
}
