package astparser

import (
	sitter "github.com/tree-sitter/go-tree-sitter"
)

// Node represents a serializable, language-agnostic Abstract Syntax Tree (AST) node.
// It wraps Tree-sitter metadata into a format that is easily consumed by Lua and JQ.
// The tree structure is preserved through the Children slice and the Fields map.
type Node struct {
	// Id is a unique numeric identifier for the node within its tree.
	Id uintptr `json:"id"`
	// Type is the grammar-defined name of the node (e.g., "class_definition", "identifier").
	Type string `json:"type"`
	// KindId is the numeric ID of the node type as defined by the Tree-sitter grammar.
	KindId uint16 `json:"kind_id"`
	// GrammarId is the internal identifier for the grammar that produced this node.
	GrammarId uint16 `json:"grammar_id"`
	// GrammarName is the human-readable name of the grammar (e.g., "python", "typescript").
	GrammarName string `json:"grammar_name"`
	// Content is the raw source text for the node.
	// To optimize memory and JSON size, content is typically only populated for leaf nodes.
	// Uses string type for JSON serialization but can be converted from bytes.
	Content string `json:"content,omitempty"`
	// ContentBytes stores a byte slice view of the original source for leaf nodes.
	// This avoids copying the source content and reduces memory usage significantly.
	ContentBytes []byte `json:"-"`
	// IsNamed indicates if the node represents a significant syntactic construct.
	IsNamed bool `json:"is_named"`
	// IsExtra indicates if the node is an auxiliary element (like comments or whitespace).
	IsExtra bool `json:"is_extra"`
	// HasError is true if this node or any of its descendants contain a parsing error.
	HasError bool `json:"has_error"`
	// IsError is true if this specific node represents a syntax error.
	IsError bool `json:"is_error"`
	// IsMissing is true if the parser expected this node but it was not found in the source.
	IsMissing bool `json:"is_missing"`
	// StartPos is the 0-indexed {row, column} coordinates where the node starts.
	StartPos [2]uint32 `json:"start_pos"`
	// EndPos is the 0-indexed {row, column} coordinates where the node ends.
	EndPos [2]uint32 `json:"end_pos"`
	// Children contains all sub-nodes in their original source order.
	Children []Node `json:"children,omitempty"`
	// Fields maps named children defined in the grammar (e.g., "name", "body") to their nodes.
	// If a field maps to multiple nodes, they are returned as a slice.
	Fields map[string]any `json:"fields,omitempty"`
}

// TransformNode recursively converts a native Tree-sitter node into a serializable Node.
// It maps Tree-sitter's field system into the Fields map and captures raw source
// for leaf nodes. This function is core of cogeni's AST transformation pipeline.
func TransformNode(n *sitter.Node, source []byte) Node {
	// Pre-allocate slices with known capacity to reduce allocations
	childCount := n.ChildCount()
	serializedChildren := make([]Node, 0, childCount/2) // Estimate ~50% are fields
	fields := make(map[string]any, 8)                   // Pre-allocate for common field count

	for i := uint(0); i < childCount; i++ {
		child := n.Child(i)
		if child == nil {
			continue
		}

		fieldName := n.FieldNameForChild(uint32(i))
		transformed := TransformNode(child, source)

		if fieldName != "" {
			if existing, ok := fields[fieldName]; ok {
				if slice, ok := existing.([]Node); ok {
					fields[fieldName] = append(slice, transformed)
				} else {
					fields[fieldName] = []Node{existing.(Node), transformed}
				}
			} else {
				fields[fieldName] = transformed
			}
		} else {
			serializedChildren = append(serializedChildren, transformed)
		}
	}

	node := Node{
		Id:          n.Id(),
		Type:        n.Kind(),
		KindId:      n.KindId(),
		GrammarId:   n.GrammarId(),
		GrammarName: n.GrammarName(),
		IsNamed:     n.IsNamed(),
		IsExtra:     n.IsExtra(),
		HasError:    n.HasError(),
		IsError:     n.IsError(),
		IsMissing:   n.IsMissing(),
		StartPos: [2]uint32{
			uint32(n.StartPosition().Row),
			uint32(n.StartPosition().Column),
		},
		EndPos: [2]uint32{
			uint32(n.EndPosition().Row),
			uint32(n.EndPosition().Column),
		},
		Children: serializedChildren,
		Fields:   fields,
	}

	// Only capture content for leaf nodes to keep the JSON size manageable.
	// Use byte slice view to avoid copying content, reducing memory usage.
	if len(serializedChildren) == 0 && len(fields) == 0 {
		start := n.StartByte()
		end := n.EndByte()
		node.ContentBytes = source[start:end]
		// Also set Content string for JSON serialization (computed lazily when needed)
		node.Content = string(node.ContentBytes)
	}

	return node
}
