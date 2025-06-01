// model/inline.go
// package model defines the abstract syntax tree (AST) nodes for EditML.
package model

// EditType represents the type of an inline edit operation.
type EditType string

// Constants for the different types of inline edits.
const (
	EditTypeAddition  EditType = "addition"
	EditTypeDeletion  EditType = "deletion"
	EditTypeComment   EditType = "comment"
	EditTypeHighlight EditType = "highlight"
)

// InlineEditNode represents an inline editorial change, such as an addition,
// deletion, comment, or highlight.
type InlineEditNode struct {
	EditType EditType // The type of edit (addition, deletion, etc.).
	Content  string   // The textual content of the edit (unescaped).
	EditorID string   // Optional: A short alphanumeric string identifying the editor.
}

// IsNode marks InlineEditNode as implementing the Node interface.
func (ien InlineEditNode) IsNode() {}
