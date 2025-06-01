// api.go
// package editml defines the public API for parsing and transforming EditML documents.
package editml

import (
	"fmt"
	"github.com/verkaro/editml-go/model"
	"github.com/verkaro/editml-go/parser"
	"github.com/verkaro/editml-go/transformer"
)

// IssueSeverity defines the severity level of an issue encountered during
// parsing or transformation.
type IssueSeverity string

// Constants for issue severity levels.
const (
	SeverityError   IssueSeverity = "error"
	SeverityWarning IssueSeverity = "warning"
)

// Issue represents an error or warning encountered during processing.
type Issue struct {
	Message  string        // A human-readable description of the issue.
	Line     int           // The line number where the issue occurred (1-based, if available).
	Column   int           // The column number where the issue occurred (1-based, if available, optional).
	Severity IssueSeverity // The severity of the issue (error or warning).
}

// Parse processes the input EditML string and returns a slice of nodes
// representing the document structure (Abstract Syntax Tree - AST),
// along with any parsing issues encountered.
//
// For the MVP, the implementation is adapted from the Backburner POC's parser.
func Parse(inputText string) (nodes []model.Node, issues []Issue) {
	// Step 1: Preprocess to remove debug comments (initially line comments).
	textWithoutDebugComments := parser.SkipDebugComments(inputText)

	// Step 2: Parse the processed text into nodes.
	parsedNodes, err := parser.ParseEditMLToNodes(textWithoutDebugComments)

	// Initialize issues slice.
	currentIssues := []Issue{}

	if err != nil {
		// For MVP, a critical error from ParseEditMLToNodes becomes a single Issue.
		currentIssues = append(currentIssues, Issue{
			Message:  fmt.Sprintf("Parsing error: %v", err),
			Line:     0, // Placeholder
			Column:   0, // Placeholder
			Severity: SeverityError,
		})
		return parsedNodes, currentIssues
	}
	return parsedNodes, currentIssues
}

// TransformCleanView takes a slice of nodes (AST) and applies transformations
// to produce a "Clean View" string. The Clean View reflects the editor's
// intended reading experience: additions are applied, deletions and comments
// are removed, highlights become plain text, and structural edits (moves/copies)
// are resolved. It also returns any issues encountered during transformation.
//
// For the MVP, the implementation is adapted from the Backburner POC's transformer.
func TransformCleanView(nodes []model.Node) (outputText string, issues []Issue) {
	// Call the internal transformation logic.
	transformedText, err := transformer.TransformToCleanView(nodes) // [cite: editML-code/transformer/transformer.go] (concept)

	currentIssues := []Issue{}
	if err != nil {
		// For MVP, a critical error from TransformToCleanView becomes a single Issue.
		currentIssues = append(currentIssues, Issue{
			Message:  fmt.Sprintf("Transformation error: %v", err),
			Line:     0, // Placeholder, transformation errors are often structural, not line-specific
			Column:   0, // Placeholder
			Severity: SeverityError,
		})
		// Even if there's an error, we might have partially transformed text (e.g. with error messages embedded).
		// Or, if the error is fatal (like duplicate source tag), transformedText might be empty.
		return transformedText, currentIssues
	}

	return transformedText, currentIssues
}
