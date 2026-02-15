package tests

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/lucasfabre/codegen/src/astparser"
)

func TestParseTypeScript(t *testing.T) {
	// Create a temporary directory for grammars
	tmpGrammarDir, err := os.MkdirTemp("", "astparser-test-grammars-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpGrammarDir) }()

	parser, err := astparser.New(tmpGrammarDir)
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
	// Create a temporary directory for grammars
	tmpGrammarDir, err := os.MkdirTemp("", "astparser-test-grammars-go-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpGrammarDir) }()

	parser, err := astparser.New(tmpGrammarDir)
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
