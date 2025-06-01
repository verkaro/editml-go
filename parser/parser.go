// parser/parser.go
// package parser provides functionality to parse EditML text into an AST.
package parser

import (
	"regexp"
	"sort"
	"strings"

	"github.com/verkaro/editml-go/model"
)

// Regex for Editor ID: 1-5 alphanumeric characters (Spec 3.3.2)
const editorIDPattern = `([a-zA-Z0-9]{1,5})`

// Inline Edit Regexes (capturing content and optional editor ID)
// Added (?s) to make . match newline characters for multiline content.
var (
	addRegex       = regexp.MustCompile(`\{\+((?s).*?)\+(?:` + editorIDPattern + `)?\}`)
	delRegex       = regexp.MustCompile(`\{\-((?s).*?)\-(?:` + editorIDPattern + `)?\}`)
	commentRegex   = regexp.MustCompile(`\{>((?s).*?)<(?:` + editorIDPattern + `)?\}`)
	highlightRegex = regexp.MustCompile(`\{=((?s).*?)=(?:` + editorIDPattern + `)?\}`)
)

// Structural Edit Regexes
// TAG pattern: Alphanumeric (Spec 3.4.1, 3.4.2). The original POC code used [a-z0-9]{5}.
// The spec is more general. We'll use a general alphanumeric pattern here.
// If a strict length is desired for tags, this pattern can be adjusted.
const structuralTagPattern = `([a-zA-Z0-9]+)`

var (
	// Move Source Regex (shorthands: move, mv, m)
	// Group 1: Operation keyword (move|mv|m)
	// Group 2: BlockContent (non-greedy, multiline with (?s))
	// Group 3: TAG
	moveSourceRegex = regexp.MustCompile(`\{(move|mv|m)~((?s).*?)~` + structuralTagPattern + `\}`)

	// Move Target Regex (shorthands: move, mv, m)
	// Group 1: Operation keyword (move|mv|m)
	// Group 2: TAG
	moveTargetRegex = regexp.MustCompile(`\{(move|mv|m):` + structuralTagPattern + `\}`)

	// Copy Source Regex (shorthands: copy, cp, c)
	// Group 1: Operation keyword (copy|cp|c)
	// Group 2: BlockContent (non-greedy, multiline with (?s))
	// Group 3: TAG
	copySourceRegex = regexp.MustCompile(`\{(copy|cp|c)~((?s).*?)~` + structuralTagPattern + `\}`)

	// Copy Target Regex (shorthands: copy, cp, c)
	// Group 1: Operation keyword (copy|cp|c)
	// Group 2: TAG
	copyTargetRegex = regexp.MustCompile(`\{(copy|cp|c):` + structuralTagPattern + `\}`)
)

// unescapeInlineContent processes escape sequences for inline edit content.
// Handles general escapes \\, \{, \} and context-specific operator escapes.
// (Spec 3.1, 3.3.1)
func unescapeInlineContent(content string, editType model.EditType) string {
	// General escapes first
	content = strings.ReplaceAll(content, "\\\\", "\\") // Backslash
	content = strings.ReplaceAll(content, "\\{", "{")   // Open curly
	content = strings.ReplaceAll(content, "\\}", "}")   // Close curly

	// Context-specific closing operator escapes
	switch editType {
	case model.EditTypeAddition:
		content = strings.ReplaceAll(content, "\\+", "+")
	case model.EditTypeDeletion:
		content = strings.ReplaceAll(content, "\\-", "-")
	case model.EditTypeComment:
		content = strings.ReplaceAll(content, "\\<", "<")
	case model.EditTypeHighlight:
		content = strings.ReplaceAll(content, "\\=", "=")
	}
	// Note: Spec 3.1 lists more characters that must be escaped in certain contexts
	// (e.g., \~, \%, \[, \]). A more comprehensive unescaper might be needed
	// in future iterations if these are critical within inline content.
	return content
}

// unescapeStructuralBlockContent handles general backslash escapes and then \~ -> ~
// for structural block content. (Spec 3.1, 3.4.1)
func unescapeStructuralBlockContent(content string) string {
	content = strings.ReplaceAll(content, "\\\\", "\\") // General unescape for literal backslashes
	content = strings.ReplaceAll(content, "\\~", "~")   // Specific unescape for literal tildes
	// Note: Spec 3.1 lists other characters. If they can appear escaped in structural content
	// and need unescaping, they might need handling here too (e.g. \{, \}).
	// However, spec 3.4.1 only explicitly mentions \~ for block content.
	return content
}

// genericMatch is a helper struct for collecting all types of identified EditML constructs.
type genericMatch struct {
	startIndex int
	endIndex   int
	node       model.Node
	// Future: could add 'priority' or 'level' for more complex overlap resolution
}

// parseEditMLToNodes is the main internal parsing function. It takes text
// (assumed to have debug comments already stripped) and returns a slice of nodes
// and any critical errors encountered during this phase.
// This function is unexported and will be called by the public editml.Parse().
func ParseEditMLToNodes(input string) ([]model.Node, error) { // Changed from ParseToNodes to parseEditMLToNodes
	var allMatches []genericMatch
	var issues []error // For collecting critical parsing errors

	// --- 1. Find Inline Addition Matches ---
	addIndices := addRegex.FindAllStringSubmatchIndex(input, -1)
	for _, m := range addIndices {
		content := input[m[2]:m[3]] // Group 1 (index 2,3) is the content
		editorID := ""
		if m[4] != -1 && m[5] != -1 { // Group 2 (index 4,5) is the editorID
			editorID = input[m[4]:m[5]]
		}
		allMatches = append(allMatches, genericMatch{
			startIndex: m[0], endIndex: m[1],
			node: model.InlineEditNode{
				EditType: model.EditTypeAddition,
				Content:  unescapeInlineContent(content, model.EditTypeAddition),
				EditorID: editorID,
			},
		})
	}

	// --- 2. Find Inline Deletion Matches ---
	delIndices := delRegex.FindAllStringSubmatchIndex(input, -1)
	for _, m := range delIndices {
		content := input[m[2]:m[3]]
		editorID := ""
		if m[4] != -1 && m[5] != -1 {
			editorID = input[m[4]:m[5]]
		}
		allMatches = append(allMatches, genericMatch{
			startIndex: m[0], endIndex: m[1],
			node: model.InlineEditNode{
				EditType: model.EditTypeDeletion,
				Content:  unescapeInlineContent(content, model.EditTypeDeletion),
				EditorID: editorID,
			},
		})
	}

	// --- 3. Find Inline Comment Matches ---
	commentIndices := commentRegex.FindAllStringSubmatchIndex(input, -1)
	for _, m := range commentIndices {
		content := input[m[2]:m[3]]
		editorID := ""
		if m[4] != -1 && m[5] != -1 {
			editorID = input[m[4]:m[5]]
		}
		allMatches = append(allMatches, genericMatch{
			startIndex: m[0], endIndex: m[1],
			node: model.InlineEditNode{
				EditType: model.EditTypeComment,
				Content:  unescapeInlineContent(content, model.EditTypeComment),
				EditorID: editorID,
			},
		})
	}

	// --- 4. Find Inline Highlight Matches ---
	highlightIndices := highlightRegex.FindAllStringSubmatchIndex(input, -1)
	for _, m := range highlightIndices {
		content := input[m[2]:m[3]]
		editorID := ""
		if m[4] != -1 && m[5] != -1 {
			editorID = input[m[4]:m[5]]
		}
		allMatches = append(allMatches, genericMatch{
			startIndex: m[0], endIndex: m[1],
			node: model.InlineEditNode{
				EditType: model.EditTypeHighlight,
				Content:  unescapeInlineContent(content, model.EditTypeHighlight),
				EditorID: editorID,
			},
		})
	}

	// --- 5. Find Move Source Matches ---
	moveSourceMatches := moveSourceRegex.FindAllStringSubmatchIndex(input, -1)
	for _, m := range moveSourceMatches {
		// m[0]:m[1] is full match; m[2]:m[3] is op keyword; m[4]:m[5] is BlockContent; m[6]:m[7] is TAG
		rawBlockContent := input[m[4]:m[5]]
		allMatches = append(allMatches, genericMatch{
			startIndex: m[0], endIndex: m[1],
			node: model.StructuralSourceNode{
				Operation:    model.OperationMove, // Normalized
				BlockContent: unescapeStructuralBlockContent(rawBlockContent),
				Tag:          input[m[6]:m[7]],
			},
		})
	}

	// --- 6. Find Move Target Matches ---
	moveTargetMatches := moveTargetRegex.FindAllStringSubmatchIndex(input, -1)
	for _, m := range moveTargetMatches {
		// m[0]:m[1] is full match; m[2]:m[3] is op keyword; m[4]:m[5] is TAG
		allMatches = append(allMatches, genericMatch{
			startIndex: m[0], endIndex: m[1],
			node: model.StructuralTargetNode{
				Operation: model.OperationMove, // Normalized
				Tag:       input[m[4]:m[5]],
			},
		})
	}

	// --- 7. Find Copy Source Matches ---
	copySourceMatches := copySourceRegex.FindAllStringSubmatchIndex(input, -1)
	for _, m := range copySourceMatches {
		rawBlockContent := input[m[4]:m[5]]
		allMatches = append(allMatches, genericMatch{
			startIndex: m[0], endIndex: m[1],
			node: model.StructuralSourceNode{
				Operation:    model.OperationCopy, // Normalized
				BlockContent: unescapeStructuralBlockContent(rawBlockContent),
				Tag:          input[m[6]:m[7]],
			},
		})
	}

	// --- 8. Find Copy Target Matches ---
	copyTargetMatches := copyTargetRegex.FindAllStringSubmatchIndex(input, -1)
	for _, m := range copyTargetMatches {
		allMatches = append(allMatches, genericMatch{
			startIndex: m[0], endIndex: m[1],
			node: model.StructuralTargetNode{
				Operation: model.OperationCopy, // Normalized
				Tag:       input[m[4]:m[5]],
			},
		})
	}

	// --- 9. Sort all found matches by their start index ---
	// This is crucial for correctly interleaving TextNodes.
	sort.Slice(allMatches, func(i, j int) bool {
		if allMatches[i].startIndex == allMatches[j].startIndex {
			// If start indices are the same, a longer match might be preferred.
			// This basic sort is usually okay for non-overlapping or clearly distinct matches.
			// The spec's non-nesting rules (3.3.4, 3.4.3) are not fully enforced by this
			// regex-based collection + sort alone. A more advanced parser would handle this hierarchically.
			// For MVP, this matches the POC's behavior.
			return allMatches[i].endIndex > allMatches[j].endIndex // Prefer longer match if starts are same
		}
		return allMatches[i].startIndex < allMatches[j].startIndex
	})

	// --- 10. Iterate through sorted matches and interleave TextNodes ---
	var nodes []model.Node
	lastIndex := 0
	for _, match := range allMatches {
		// Basic overlap detection: if a match starts before the last one ended,
		// it's an overlap. The POC's approach was to skip.
		// A more robust parser would use a state machine or recursive descent to avoid
		// generating overlapping matches in the first place or to implement "Graceful Failure" (Spec 4.5)
		// and non-nesting rules (Spec 3.3.4, 3.4.3) more precisely.
		if match.startIndex < lastIndex {
			// Overlap detected. For MVP, we skip this match to mimic POC behavior.
			// Future: This could be an 'issue' or handled by a more sophisticated parsing strategy.
			// For now, this prioritizes the "earlier" or "outer" match if sorted correctly.
			continue
		}

		// Add preceding text as a TextNode
		if match.startIndex > lastIndex {
			nodes = append(nodes, model.TextNode{Text: input[lastIndex:match.startIndex]})
		}
		// Add the matched EditML node
		nodes = append(nodes, match.node)
		lastIndex = match.endIndex
	}

	// Add any remaining text after the last match
	if lastIndex < len(input) {
		nodes = append(nodes, model.TextNode{Text: input[lastIndex:]})
	}

	// Handle empty input: if input is empty and no nodes were produced, return empty slice, no error.
	if input == "" && len(nodes) == 0 {
		return []model.Node{}, nil
	}

	// If any critical errors were collected, return them. For MVP, this 'issues' list is basic.
	if len(issues) > 0 {
		// For now, we'll just return the first error if multiple occurred.
		// A more robust system would return all 'issues'.
		return nodes, issues[0]
	}

	return nodes, nil
}
