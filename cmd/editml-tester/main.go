// cmd/editml-tester/main.go
// This is a simple CLI tool to test the editml API.
package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/verkaro/editml-go"
	"github.com/verkaro/editml-go/model" // For printing model.Node details in debug mode
)

func main() {
	// Define a debug flag
	debug := flag.Bool("debug", false, "Enable debug output (prints AST and issues)")
	flag.Parse()

	// Read input from stdin
	// fmt.Fprintln(os.Stderr, "Enter EditML text (press Ctrl+D to end input):") // Prompt
	inputBytes, err := io.ReadAll(os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading from stdin: %v\n", err)
		os.Exit(1)
	}
	inputText := string(inputBytes)

	if *debug {
		fmt.Println("--- Input Text ---")
		// To ensure multiline input is clearly demarcated, especially if it's short
		// and doesn't end with a newline from the pipe.
		// Using a simple print here. For complex inputs, might need more careful formatting.
		printWithLineNumbers(inputText)
		fmt.Println("--- End Input Text ---")
		fmt.Println()
	}

	// Call the editml API's Parse function
	nodes, parseIssues := editml.Parse(inputText)

	if *debug {
		fmt.Println("--- Parsing Results (AST) ---")
		if len(nodes) > 0 {
			for i, node := range nodes {
				fmt.Printf("Node %d: %s\n", i+1, formatNode(node))
			}
		} else {
			fmt.Println("(No nodes parsed)")
		}
		fmt.Println("--- End AST ---")
		fmt.Println()

		fmt.Println("--- Parsing Issues ---")
		if len(parseIssues) > 0 {
			for _, issue := range parseIssues {
				fmt.Printf("[%s] L%d:%d %s\n", issue.Severity, issue.Line, issue.Column, issue.Message)
			}
		} else {
			fmt.Println("(None)")
		}
		fmt.Println("--- End Parsing Issues ---")
		fmt.Println()
	}

	// Call the editml API's TransformCleanView function
	outputText, transformIssues := editml.TransformCleanView(nodes)

	if *debug {
		fmt.Println("--- Transformation Issues ---")
		if len(transformIssues) > 0 {
			for _, issue := range transformIssues {
				fmt.Printf("[%s] L%d:%d %s\n", issue.Severity, issue.Line, issue.Column, issue.Message)
			}
		} else {
			fmt.Println("(None)")
		}
		fmt.Println("--- End Transformation Issues ---")
		fmt.Println()

		fmt.Println("--- Final Output (Clean View) ---")
	}

	// Print the final transformed output
	// If outputText is empty and there were no errors, it means the transformation resulted in an empty string.
	// If there were errors, outputText might contain error messages or be partially transformed.
	fmt.Print(outputText) // Use Print to avoid an extra newline if outputText already has one.
	if *debug && len(outputText) > 0 && !strings.HasSuffix(outputText, "\n") {
		fmt.Println() // Add a newline in debug mode if output didn't have one, for cleaner debug log.
	}

	if *debug {
		fmt.Println("--- End Final Output ---")
	}

	// Exit with an error code if there were any "error" severity issues
	hasErrors := false
	for _, issue := range parseIssues {
		if issue.Severity == editml.SeverityError {
			hasErrors = true
			break
		}
	}
	if !hasErrors {
		for _, issue := range transformIssues {
			if issue.Severity == editml.SeverityError {
				hasErrors = true
				break
			}
		}
	}

	if hasErrors {
		if !*debug { // If not in debug, errors might not have been printed yet
			fmt.Fprintln(os.Stderr, "Errors occurred during processing. Run with --debug for details.")
		}
		os.Exit(1)
	}
}

// formatNode provides a string representation of a model.Node for debug printing.
func formatNode(node model.Node) string {
	switch n := node.(type) {
	case model.TextNode:
		// Truncate long text for readability in debug output
		textContent := n.Text
		if len(textContent) > 50 {
			textContent = textContent[:47] + "..."
		}
		return fmt.Sprintf("TextNode (Text: %q)", textContent)
	case model.InlineEditNode:
		return fmt.Sprintf("InlineEditNode (Type: %s, Content: %q, EditorID: %q)", n.EditType, n.Content, n.EditorID)
	case model.StructuralSourceNode:
		return fmt.Sprintf("StructuralSourceNode (Operation: %s, Tag: %q, BlockContent: %q)", n.Operation, n.Tag, n.BlockContent)
	case model.StructuralTargetNode:
		return fmt.Sprintf("StructuralTargetNode (Operation: %s, Tag: %q)", n.Operation, n.Tag)
	default:
		return fmt.Sprintf("Unknown Node Type: %T", n)
	}
}

// printWithLineNumbers prints the given text with line numbers, useful for debugging input.
func printWithLineNumbers(text string) {
	scanner := bufio.NewScanner(strings.NewReader(text))
	lineNumber := 1
	for scanner.Scan() {
		fmt.Printf("%4d: %s\n", lineNumber, scanner.Text())
		lineNumber++
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "Error scanning text for line numbers: %v\n", err)
	}
}
