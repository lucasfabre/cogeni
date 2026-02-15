package astparser

import (
	"fmt"

	sitter "github.com/tree-sitter/go-tree-sitter"
)

// Parser is the high-level orchestrator for parsing source code into a detailed AST.
// It manages grammar loading and the underlying tree-sitter parser lifecycle.
type Parser struct {
	grammarManager *GrammarManager
}

// New creates a new Parser instance.
// grammarDir is the directory where grammar shared libraries will be stored,
// and source code will be downloaded if necessary.
func New(grammarDir string) (*Parser, error) {
	gm, err := NewGrammarManager(grammarDir)
	if err != nil {
		return nil, fmt.Errorf("failed to create grammar manager: %w", err)
	}
	return &Parser{grammarManager: gm}, nil
}

// RegisterGrammar allows providing a custom source and build configuration for a grammar.
// This overrides the default lookup from the tree-sitter organization.
func (p *Parser) RegisterGrammar(name, url string, opts GrammarOptions) {
	p.grammarManager.RegisterGrammar(name, url, opts)
}

// Parse converts source code into a serializable Node structure.
// It handles grammar loading (downloading/compiling if needed), initializes
// the Tree-sitter parser, and performs the recursive transformation from
// native sitter.Node to astparser.Node.
func (p *Parser) Parse(langName string, source []byte) (Node, error) {
	lang, err := p.grammarManager.GetLanguage(langName)
	if err != nil {
		return Node{}, fmt.Errorf("failed to load language '%s': %w", langName, err)
	}

	parser := sitter.NewParser()
	defer parser.Close()

	if err := parser.SetLanguage(lang); err != nil {
		return Node{}, fmt.Errorf("failed to set language '%s': %w", langName, err)
	}

	tree := parser.Parse(source, nil)
	defer tree.Close()

	if tree == nil {
		return Node{}, fmt.Errorf("parsing failed")
	}

	root := tree.RootNode()
	if root == nil {
		return Node{}, fmt.Errorf("parsing produced an empty tree")
	}

	return TransformNode(root, source), nil
}
