// model/structural.go
// package model defines the abstract syntax tree (AST) nodes for EditML.
package model

// Constants for structural operation types.
const (
	OperationMove = "move"
	OperationCopy = "copy"
)

// StructuralSourceNode represents a block of text marked for a structural
// operation like move or copy.
// Example: {move~block content~TAG} or {copy~block content~TAG}
type StructuralSourceNode struct {
	Operation    string // The type of operation (e.g., "move", "copy").
	Tag          string // The unique alphanumeric identifier for this block.
	BlockContent string // The raw textual content within the tildes (unescaped).
	// For MVP, BlockContent is a string. Future iterations may parse this
	// into []Node if it can contain further EditML markup as per spec.
}

// IsNode marks StructuralSourceNode as implementing the Node interface.
func (ssn StructuralSourceNode) IsNode() {}

// StructuralTargetNode represents a target location for a structural operation.
// Example: {move:TAG} or {copy:TAG}
type StructuralTargetNode struct {
	Operation string // The type of operation (e.g., "move", "copy").
	Tag       string // The alphanumeric identifier linking to a source block.
}

// IsNode marks StructuralTargetNode as implementing the Node interface.
func (stn StructuralTargetNode) IsNode() {}
