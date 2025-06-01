// api_test.go
// package editml_test contains unit tests for the editml API.
package editml

import (
	"os"            // For reading test files
	"path/filepath" // For joining path elements
	"reflect"       // For deep equality comparison of slices and structs
	"strings"       // For string manipulations like TrimSpace
	"testing"

	"github.com/verkaro/editml-go/model"
)

// TestParseSimpleAddition tests the Parse function with a basic addition.
func TestParseSimpleAddition(t *testing.T) {
	inputText := "{+added text+ws}"
	expectedNodes := []model.Node{
		model.InlineEditNode{
			EditType: model.EditTypeAddition,
			Content:  "added text",
			EditorID: "ws",
		},
	}
	// For MVP, we expect no issues from this simple valid input.
	expectedIssues := []Issue{}

	actualNodes, actualIssues := Parse(inputText)

	if !reflect.DeepEqual(actualNodes, expectedNodes) {
		t.Errorf("Parse(%q) nodes = %v, want %v", inputText, actualNodes, expectedNodes)
	}

	if !reflect.DeepEqual(actualIssues, expectedIssues) {
		t.Errorf("Parse(%q) issues = %v, want %v", inputText, actualIssues, expectedIssues)
	}
}

// TestParseTextAndAddition tests parsing a mix of plain text and an addition.
func TestParseTextAndAddition(t *testing.T) {
	inputText := "Hello {+World+}"
	expectedNodes := []model.Node{
		model.TextNode{Text: "Hello "},
		model.InlineEditNode{EditType: model.EditTypeAddition, Content: "World", EditorID: ""},
	}
	expectedIssues := []Issue{}

	actualNodes, actualIssues := Parse(inputText)

	if !reflect.DeepEqual(actualNodes, expectedNodes) {
		t.Errorf("Parse(%q) nodes = %v, want %v", inputText, actualNodes, expectedNodes)
	}
	if !reflect.DeepEqual(actualIssues, expectedIssues) {
		t.Errorf("Parse(%q) issues = %v, want %v", inputText, actualIssues, expectedIssues)
	}
}

// TestTransformSimpleAddition tests TransformCleanView with a parsed addition.
func TestTransformSimpleAddition(t *testing.T) {
	inputNodes := []model.Node{
		model.InlineEditNode{
			EditType: model.EditTypeAddition,
			Content:  "added text",
			EditorID: "ws",
		},
	}
	expectedOutput := "added text"
	// For MVP, we expect no issues from this simple valid transformation.
	expectedIssues := []Issue{}

	actualOutput, actualIssues := TransformCleanView(inputNodes)

	if actualOutput != expectedOutput {
		t.Errorf("TransformCleanView for simple addition: output = %q, want %q", actualOutput, expectedOutput)
	}

	if !reflect.DeepEqual(actualIssues, expectedIssues) {
		t.Errorf("TransformCleanView for simple addition: issues = %v, want %v", actualIssues, expectedIssues)
	}
}

// TestTransformTextAndDeletion tests a mix of text and a deletion.
func TestTransformTextAndDeletion(t *testing.T) {
	// Simulate parsing "Hello {-World-}"
	inputNodes := []model.Node{
		model.TextNode{Text: "Hello "},
		model.InlineEditNode{EditType: model.EditTypeDeletion, Content: "World", EditorID: ""},
	}
	expectedOutput := "Hello " // Deletion content is removed
	expectedIssues := []Issue{}

	actualOutput, actualIssues := TransformCleanView(inputNodes)

	if actualOutput != expectedOutput {
		t.Errorf("TransformCleanView for text and deletion: output = %q, want %q", actualOutput, expectedOutput)
	}
	if !reflect.DeepEqual(actualIssues, expectedIssues) {
		t.Errorf("TransformCleanView for text and deletion: issues = %v, want %v", actualIssues, expectedIssues)
	}
}

// TestParseAndTransformIntegration is a simple integration test.
func TestParseAndTransformIntegration(t *testing.T) {
	inputText := "This is {+an addition+} and this is {-a deletion-}."
	expectedOutput := "This is an addition and this is ."

	nodes, parseIssues := Parse(inputText)
	if len(parseIssues) > 0 {
		t.Fatalf("Parse(%q) returned unexpected issues: %v", inputText, parseIssues)
	}

	output, transformIssues := TransformCleanView(nodes)
	if len(transformIssues) > 0 {
		t.Fatalf("TransformCleanView returned unexpected issues: %v", transformIssues)
	}

	if output != expectedOutput {
		t.Errorf("ParseAndTransformIntegration: output = %q, want %q", output, expectedOutput)
	}
}

// TestParseAndTransformMultilineFile tests parsing and transforming the multiline.md test file.
func TestParseAndTransformMultilineFile(t *testing.T) {
	// Read the content of testdata/multiline.md
	filePath := filepath.Join("testdata", "multiline.md")
	inputBytes, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read test file %s: %v", filePath, err)
	}
	inputText := string(inputBytes)

	// This is the "Final Output (Clean View)" based on the AST interpretation and debug output,
	// corrected to have:
	// 1. Two blank lines after "This is a test document for multiline EditML features."
	// 2. Three blank lines after "Now for structural edits...".
	expectedMultilineOutput := `This is a test document for multiline EditML features.


First, some multiline inline additions:
This is a simple
addition spanning
multiple lines. an example.

Another one without an ID:
This is another
multiline addition.
It even has a blank line

in the middle..

Multiline deletions:
This text is  and should be removed.

And one more:
.

Multiline comments:
Here is some text  that we want to keep.

A comment with an escaped closing operator:
.

Multiline highlights:
Let's emphasize this
important section
of text. for review.

Another highlight:
This part
is also
highlighted..

Now for structural edits with multiline content.



Some text in between.

This is a block
to be copied.
Line 1
Line 2
Line 3

More text.

And here are the targets:
The moved text should appear here: This is a block of text
that is intended to be moved.
It has multiple lines.
It even contains an escaped tilde: ~ here.
.

The copied text should appear here: This is a block
to be copied.
Line 1
Line 2
Line 3.
And also here: This is a block
to be copied.
Line 1
Line 2
Line 3.

Mixing it up:
This is an added line
that also contains {-a nested
deletion (which should be literal text per spec 3.3.4)-}
and more added text..

A structural move with inline edits inside its content:


The target for the mixed content move:
This block will be moved.
It contains added text within the block
and also some .
This should all move together.


End of multiline tests.`

	nodes, parseIssues := Parse(inputText)
	if len(parseIssues) > 0 {
		t.Fatalf("Parse for multiline.md returned unexpected issues: %v", parseIssues)
	}

	actualOutput, transformIssues := TransformCleanView(nodes)
	if len(transformIssues) > 0 {
		t.Fatalf("TransformCleanView for multiline.md returned unexpected issues: %v", transformIssues)
	}

	// Primary comparison: exact match.
	if actualOutput != expectedMultilineOutput {
		// Fallback comparison if exact match fails, to see if it's just trailing whitespace.
		if strings.TrimSpace(actualOutput) != strings.TrimSpace(expectedMultilineOutput) {
			t.Errorf("ParseAndTransformMultilineFile: output (trimmed) does not match expected (trimmed).\nExpected:\n%s\n\nActual:\n%s", expectedMultilineOutput, actualOutput)
		} else {
			// Content is the same after trimming, so it's likely a subtle leading/trailing whitespace or newline difference.
			t.Errorf("ParseAndTransformMultilineFile: output matches expected when trimmed, but differs in exact whitespace/newlines.\nExpected:\n%s\n\nActual:\n%s", expectedMultilineOutput, actualOutput)
		}
	}
}
