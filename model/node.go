// package model defines the abstract syntax tree (AST) nodes for EditML.
package model

// Node is the interface implemented by all AST node types.
type Node interface {
	IsNode() // Marker method to ensure type safety.
}

// TextNode represents a block of plain text in the document.
type TextNode struct {
	Text string
}

// IsNode marks TextNode as implementing the Node interface.
func (tn TextNode) IsNode() {}
