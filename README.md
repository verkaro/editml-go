# EditML Go Library

## Overview

This Go library provides tools for parsing and transforming EditML-formatted text. EditML is a plaintext-compatible markup language designed to express editorial intent within prose documents, supporting inline micro-edits and block-level structural operations.

The primary purpose of this library is to make editorial revisions transparent, easily auditable, human-readable, and interoperable across tooling. It aims to parse EditML into an Abstract Syntax Tree (AST) and then transform this AST into different views, with an initial focus on a "Clean View" for direct readability.

This library was initially developed for use in projects like Pivot3, an interactive fiction authoring tool.

## Current Status

  * **MVP (Minimum Viable Product)**
  * Supports a functional subset of [EditML Specification v2.5](docs/EditML-Spec-v2.5.md).
  * Primary output profile implemented: "Clean View".

## Features (MVP)

  * Parsing of EditML line debug comments (`%% ...`).
  * Parsing of inline edits:
      * Additions (`{+...+}`)
      * Deletions (`{-...-}`)
      * Comments (`{>...<}`)
      * Highlights (`{=...=}`)
      * Supports multiline content and optional editor IDs for inline edits.
  * Parsing of structural edits:
      * Move operations (`{move~...~TAG}`, `{move:TAG}`)
      * Copy operations (`{copy~...~TAG}`, `{copy:TAG}`)
      * Supports shorthand keywords (e.g., `mv`, `cp`).
  * Transformation to "Clean View":
      * Additions are applied directly into the text.
      * Deletions and comments are omitted from the output.
      * Highlights are rendered as their plain text content.
      * Structural edits (moves and copies) are resolved.
  * Basic error and issue reporting via the `editml.Issue` struct.

## Prerequisites

  * Go 1.21+.

## Installation

```bash
go get [https://github.com/verkaro/editml-go](https://github.com/verkaro/editml-go)
````

## Basic Usage (Go Example)

```go
package main

import (
    "fmt"
    "log"

    "github.com/verkaro/editml-go"
)

func main() {
    inputText := "Hello {+World+}! This is {-not seen-}."
    
    // Parse the EditML input
    nodes, parseIssues := editml.Parse(inputText)

    // Check for parsing issues
    if len(parseIssues) > 0 {
        for _, issue := range parseIssues {
            if issue.Severity == editml.SeverityError {
                log.Printf("Parse Error: %s (L%d:%d)", issue.Message, issue.Line, issue.Column)
            } else if issue.Severity == editml.SeverityWarning {
                log.Printf("Parse Warning: %s (L%d:%d)", issue.Message, issue.Line, issue.Column)
            }
        }
        // Decide if to proceed if only warnings, or stop on error
        // For example, if any error issue exists:
        // for _, issue := range parseIssues { if issue.Severity == editml.SeverityError { return } }
    }

    // Transform the AST to a Clean View
    cleanText, transformIssues := editml.TransformCleanView(nodes)

    // Check for transformation issues
    if len(transformIssues) > 0 {
        for _, issue := range transformIssues {
             if issue.Severity == editml.SeverityError {
                log.Printf("Transform Error: %s (L%d:%d)", issue.Message, issue.Line, issue.Column)
            } else if issue.Severity == editml.SeverityWarning {
                log.Printf("Transform Warning: %s (L%d:%d)", issue.Message, issue.Line, issue.Column)
            }
        }
    }

    fmt.Println(cleanText) // Expected output for the example: "Hello World! This is ."
}

```

## `editml-tester` CLI Tool

This repository includes a command-line tool, `editml-tester`, for hands-on testing and debugging of the `editml` library.

### How to Build

From the root directory of the `editml` repository:

```bash
go build -o editml-tester ./cmd/editml-tester/
```

### How to Run

```bash
# Basic usage (reads from stdin)
./editml-tester < path/to/your/file.md

# With debug output (shows AST, issues, etc.)
./editml-tester --debug < path/to/your/file.md
```

## Directory Structure

  * `api.go`, `issue.go` (etc.): Core public API for the `editml` package.
  * `model/`: Defines the Abstract Syntax Tree (AST) node structures.
  * `parser/`: Internal parsing logic.
  * `transformer/`: Internal transformation logic.
  * `cmd/editml-tester/`: Source for the CLI testing tool.
  * `testdata/`: Contains sample `.md` files used for testing (e.g., `simple.md`, `multiline.md`).


## TODO List for editml Go Library

This list outlines planned future enhancements and areas for improvement for the `editml` library, building upon the initial MVP.

### Fuller EditML Specification v2.5 Compliance

* [ ] **Parse Block Debug Comments:** Implement parsing for `%%[...]%%` block comments (Spec 3.2.2).
* [ ] **Enhance Parser Robustness:**
    * [ ] **True Non-Nesting for Inline Edits:** Ensure parser strictly adheres to Spec 3.3.4 for complex nested scenarios. (Current MVP handles simple cases correctly).
    * [ ] **Full "Graceful Failure":** Implement complete behavior for unknown `{...}` blocks as per Spec 4.5 (treat as single TextNode).
* [ ] **Comprehensive Validation Rules:** Implement checks during or after parsing for:
    * [ ] Unique source tags for structural edits (Spec 3.4.3).
    * [ ] No dual operation type (move/copy) for the same tag (Spec 3.4.3).
    * [ ] Other structural rule validations as defined in the specification.
* [ ] **Strict Structural Edit Execution:**
    * [ ] Enforce strict execution order: all copy operations first, then all move operations (Spec 5.1.1).
    * [ ] Implement full structural conflict resolution: abort all structural transformations on any conflict (Spec 5.1.1).
* [ ] **Improved AST for Structural Content:**
    * [ ] Modify parser so `StructuralSourceNode.BlockContent` is parsed into `[]model.Node` directly by `editml.Parse`, rather than being re-parsed by the transformer (Ref: Spec 3.4.3 allowing bbtext within bbstructure).

### Additional Transformation Profiles

* [ ] **Implement `MarkupView` Profile:** Create a transformation that preserves all EditML markup literally (Spec 5.1).
* [ ] **Implement `HTMLPreview` Profile:** Create a transformation that renders EditML with basic HTML styling for inline edits (e.g., `<ins>`, `<del>`, styled spans) (Spec 5.1, 6).

### Error and Issue Reporting

* [ ] **Precise Positional Information:** Enhance `editml.Issue` reporting to include accurate line and column numbers for all issues.
* [ ] **Granular Warnings:** Introduce more specific warning types for non-fatal deviations from the specification or best practices.

### Parser Architecture

* [ ] **Evaluate Parser Refactoring:** Consider and potentially implement a refactor of the parser to use a more traditional tokenizer + recursive descent (or similar compiler-front-end technique) approach. This could improve maintainability, robustness, and ease of adding more complex parsing rules over the current regex-based method.

### Testing

* [ ] **Expand Unit Tests:** Add more unit tests covering:
    * A wider range of edge cases for inline and structural edits.
    * Invalid EditML syntax and expected error/issue reporting.
    * All implemented transformation profiles (`MarkupView`, `HTMLPreview`).
    * Specific validation error scenarios.
* [ ] **Integration Tests:** Consider more complex integration tests that combine multiple features.

### Documentation

* [ ] **Comprehensive GoDoc Comments:** Write thorough GoDoc comments for all public API elements (functions, types, constants, fields) to enable auto-generated documentation.
* [ ] **Expand README Examples:** Add more usage examples to `README.md` as new features and profiles are implemented.

{+This project benefited from AI assistance (Google's Gemini model) during its planning and development stages.+}



