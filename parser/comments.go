// parser/comments.go
// package parser provides functionality to parse EditML text into an AST.
package parser

import (
	"bufio"
	"strings"
	"unicode"
	"unicode/utf8"
)

// skipDebugComments processes an input string and removes EditML line comments.
// An EditML line comment starts with '%%' followed by a space, tab, newline,
// end of file, or any non-alphanumeric character.
// If '%%' is immediately followed by an alphanumeric character, it's treated as literal text.
//
// For the MVP, this handles line comments. Block comment handling ('%%[...]%%')
// is a future enhancement.
func SkipDebugComments(input string) string {
	var resultLines []string
	scanner := bufio.NewScanner(strings.NewReader(input))

	for scanner.Scan() {
		line := scanner.Text()
		isCommentLine := false

		if strings.HasPrefix(line, "%%") {
			// Check if it's just "%%" or if the character after "%%" makes it a comment.
			if len(line) == 2 { // Line is exactly "%%"
				isCommentLine = true
			} else {
				// Decode the first rune after "%%"
				charAfter, _ := utf8.DecodeRuneInString(line[2:])

				// According to Spec 3.2.1:
				// "%%" must be followed by an ASCII space (U+0020), a horizontal tab (U+0009),
				// a newline character, the end of the file, or any non-alphanumeric character.
				// If %% is immediately followed by an alphanumeric character, %% is treated as literal text.
				if unicode.IsLetter(charAfter) || unicode.IsDigit(charAfter) {
					isCommentLine = false // e.g., "%%VERSION" is not a comment
				} else {
					isCommentLine = true // e.g., "%% This is a comment", "%%-not-alphanum"
				}
			}
		}

		if !isCommentLine {
			resultLines = append(resultLines, line)
		}
	}
	// Rejoin with newline. Note: This might alter original newline conventions if mixed (e.g. \r\n vs \n)
	// but is generally fine for typical text processing.
	return strings.Join(resultLines, "\n")
}
