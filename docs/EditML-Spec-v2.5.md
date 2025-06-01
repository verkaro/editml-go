# EditML Specification v2.5

## 1. Overview

EditML is a plaintext-compatible markup language designed to express editorial intent within prose documents. It supports inline micro-edits and block-level structural operations. Its primary goals are to make editorial revisions transparent, easily auditable over time, human-readable, and interoperable across tooling. This version emphasizes human-friendly syntax while maintaining machine-parsability and aiming for implementation consistency.

## 2. Core Concepts

* **Human-Readable & Writable**: Syntax should be intuitive, visually clean, and minimize cognitive load for authors working directly in plain text.
* **Machine-Parsable**: The language must be unambiguously parsable into an Abstract Syntax Tree (AST) suitable for programmatic transformation and analysis.
* **Declarative Intent**: Markup represents editorial *intent*. The application of these intents is handled by a transformation process.
* **Multiline-Safe**: Markup for both inline and structural edits can span multiple lines.
* **Attribution-Ready**: Optional author/editor initials (IDs) can be associated with inline edits.
* **Transformation-Focused**: The primary way to interact with EditML is by transforming the document into different views.

## 2.1. Supported Editorial Capabilities (Summary)
EditML allows authors to:
* Propose text to be added.
* Mark text for deletion.
* Insert inline annotations or comments.
* Highlight text for review or emphasis.
* Define blocks of text to be moved to other locations.
* Define blocks of text to be copied to one or more locations.
* Temporarily comment out sections of text or markup for debugging or review.

## 3. Syntax

### 3.1. General Escape Character
The backslash `\` is used as the escape character. To include a literal character that has special meaning in EditML syntax, precede it with a backslash. For a literal backslash, use `\\`. Parsers must resolve these escape sequences when processing content, meaning the AST node's content string should contain the literal character, not the escape sequence.

The following characters *must* be escaped with a preceding backslash `\` if they are to be treated as literal characters within contexts where they have special syntactic meaning:
* `{` (curly brace open)
* `}` (curly brace close)
* `~` (tilde)
* `%` (percent, when part of `%%`, `%%[`, or `]%%` sequences)
* `[` (square bracket open, when immediately following `%%` to form `%%[`)
* `]` (square bracket close, when immediately followed by `%%` to form `]%%`)
* `<` (less-than sign, when used as the closing operator for `{>...<}` comments and needing to appear literally within the comment's content)
* `\` (backslash itself)

*Example:* `\{` in input becomes `{` in AST content; `\~` becomes `~`; `\\` becomes `\`; `\<` within a comment's content becomes `<`.

### 3.2. Debug Comments

Debug comments are used to exclude text from EditML parsing and processing.

#### 3.2.1. Line Comments
Any line beginning with `%%` (where both `%` characters are unescaped) is treated as a line comment. The `%%` and all subsequent text on that line are ignored by the parser.
* **Condition:** For `%%` to initiate a comment, it must be followed by an ASCII space (U+0020), a horizontal tab (U+0009), a newline character, the end of the file, or any non-alphanumeric character. If `%%` is immediately followed by an alphanumeric character, `%%` is treated as literal text (e.g., `%%VERSION` is not a comment).
* **Escape:** To have a literal `%%` at the beginning of a line that *is* followed by a space/non-alphanumeric but should *not* be a comment, escape one of the percent signs (e.g., `\%% This is not a comment` or `%\%% This is not a comment`).

*Example:*


%% This is a line comment and will be ignored.
This is normal text.
%% This line starts with literal "%%".
%%VERSION is not a comment.
#### 3.2.2. Block Comments
A block comment starts with an unescaped `%%[` sequence and ends with an unescaped `]%%` sequence. All content between these delimiters, including line breaks and any EditML-like syntax, is ignored by the parser. Block comments cannot be nested.
* **Escape:** To include literal `%%[` or `]%%` sequences within the *content* of a block comment, escape the specific character that would form part of the delimiter.
    * To include `%%[` literally: use `\%%[` (escaping the first `%`) or `%%\\[` (escaping the `[`). The simpler and recommended form is `%%\\[`.
    * To include `]%%` literally: use `\]\%%` (escaping the `]`) or `]\%%` (escaping the first `%`). The simpler and recommended form is `\]\%%`.
    The principle is that the `\` escapes the immediately following character.

*Example:*


This is normal text.
%%[
This entire block,
including {+this "markup"+} and even literal ]%% (escaped),
is ignored.
And %%\[ (escaped) is also ignored.
]%%
More normal text.
### 3.3. Inline Edit Markup (bbtext)

Inline edits are enclosed in unescaped `{` and `}`. They consist of an opening operator, content, a closing operator, and an optional editor ID.

#### 3.3.1. Syntax Forms

| Type      | Syntax Pattern      | Example (No ID)    | Example (With ID `ws`) |
| :-------- | :------------------ | :----------------- | :--------------------- |
| Addition  | `{+text[+id]}`      | `{+added text+}`   | `{+added text+ws}`     |
| Deletion  | `{-text[-id]}`      | `{-deleted text-}` | `{-deleted text-ws}`   |
| Comment   | `{>text<[id]}`      | `{>my comment<}`   | `{>my comment<ws}`     |
| Highlight | `{=text=[id]}`      | `{=important=}`    | `{=important=ws}`      |

* **Content (`text`):** The actual text being added, deleted, commented on, or highlighted. Can be multiline. Literal `{`, `}`, and the specific closing operator for that edit type (e.g., `<` for comments) must be escaped (e.g., `\{`, `\}`, `\<`) if they are to appear literally within the content. Standard whitespace characters like ASCII spaces, tabs, and newlines within content are preserved and do not require escaping.
* **Operators:**
    * Addition: `+` (opening and closing)
    * Deletion: `-` (opening and closing, using hyphen-minus)
    * Comment: `>` (opening) and `<` (closing)
    * Highlight: `=` (opening and closing)
* **Editor ID (`id`):** Optional. A short alphanumeric string identifying the author/editor. If present, it immediately follows the closing content operator and precedes the final `}`.

#### 3.3.2. Editor ID Format
Editor IDs must strictly consist of alphanumeric characters (a-z, A-Z, 0-9) and typically be 1-5 characters in length. They must not contain whitespace (ASCII space or tab) or EditML delimiter/operator characters.

#### 3.3.3. Multiline Content
The content within inline edits can span multiple lines.

#### 3.3.4. Nesting
Inline edits (`bbtext`) **cannot** be nested within other inline edits. If an opening EditML inline delimiter sequence (e.g., `{+`) appears within the *content* of another inline edit, it and its content are treated as literal text (unless the opening `{` is escaped as `\{`). For example, in `{+This is {=important=} text+}`, the `{=important=}` part is literal text content of the addition.

### 3.4. Structural Edit Markup (bbstructure)

Structural edits move or copy blocks of text.

#### 3.4.1. Source Syntax
* **Move Source:** `{move~block content~TAG}`
* **Copy Source:** `{copy~block content~TAG}`
    * Shorthands for `move`: `mv`, `m`
    * Shorthands for `copy`: `cp`, `c`

* **Components:**
    1.  `{`: Opening delimiter (unescaped).
    2.  `operation_keyword`: `move`, `mv`, `m`, `copy`, `cp`, `c`. Case-sensitive.
    3.  `~`: First delimiter.
    4.  `block content`: The text to be moved/copied. Can be multiline and can contain `bbtext` inline markup. Literal `~` characters within `block content` must be escaped as `\~`. The parser must identify the correct, unescaped closing `~TAG}` sequence. Standard whitespace characters (spaces, tabs, newlines) within content are preserved and do not require escaping.
    5.  `~`: Second delimiter.
    6.  `TAG`: A unique identifier strictly consisting of alphanumeric characters (a-z, A-Z, 0-9), and thus cannot contain whitespace or EditML delimiters/operators. Case-sensitive.
    7.  `}`: Closing delimiter (unescaped).

#### 3.4.2. Target Syntax
* **Move Target:** `{move:TAG}`
* **Copy Target:** `{copy:TAG}`
    * Shorthands for `move`: `mv`, `m`
    * Shorthands for `copy`: `cp`, `c`

* **Components:**
    1.  `{`: Opening delimiter (unescaped).
    2.  `operation_keyword`: `move`, `mv`, `m`, `copy`, `cp`, `c`. Case-sensitive.
    3.  `:`: Delimiter.
    4.  `TAG`: Alphanumeric identifier linking to a source block. Must strictly consist of alphanumeric characters (a-z, A-Z, 0-9), and thus cannot contain whitespace or EditML delimiters/operators.
    5.  `}`: Closing delimiter (unescaped).

#### 3.4.3. Structural Rules
* **Tag Uniqueness for Sources:** Each `TAG` in a source block must be unique. Duplicate source tags are an error.
* **No Dual Operation Type for a Tag:** A `TAG` cannot define both a move and a copy source.
* **Unresolved Tags:** Source or target blocks can exist without their counterpart. This is not a validation error. Tools *should* provide "info" or "warning" level feedback for unresolved tags to aid authors.
* **Target Multiplicity:**
    * `copy` operations support multiple targets (multiple `{copy:TAG}` for the same `TAG`).
    * `move` operations imply a single source and should have at most one corresponding `move` target. If multiple `{move:TAG}` targets (or their shorthands) exist for the same `TAG`, this constitutes a structural conflict. Implementations *must* report this as an error. Consistent with the conflict resolution policy in Section 5.1.1, the transformation process for structural edits *should* be aborted.
* **Nesting:** `bbstructure` tags cannot be nested within other `bbstructure` tags. `bbtext` *can* be within the `block content` of a `bbstructure` source.
* **No Editor IDs:** `bbstructure` tags do not support editor IDs.

## 4. Parser Behavior

### 4.1. Dispatching Logic
The primary parsing function (e.g., `api.ProcessDocument`) determines block type:
1.  **Debug Comment:** If input matches `%%` (line) or `%%[` (block) per rules in 3.2, consume and ignore.
2.  **EditML Block:** If input starts with an unescaped `{`:
    * Examine the character(s) immediately following `{`.
    * If `+`, `-`, `=`, `>`: Dispatch to `bbtext` parser.
    * If a known structural keyword (`move`, `mv`, etc.) followed by `~` (source) or `:` (target): Dispatch to `bbstructure` parser.
    * **Graceful Failure (Unknown `{` Block):** If `{` is followed by characters not matching defined patterns, treat the block from `{` to its corresponding matching `}` as plain text. (See 4.5).
3.  **Plain Text:** Other text is plain content.

### 4.2. Tokenization
* Identify and enable skipping of debug comments.
* Recognize EditML delimiters (`{`, `}`, `~`, `:`) and operators (`+`, `-`, `=`, `>`, `<`).
* Identify operation keywords and tags (alphanumeric sequences).
* **Whitespace in Markup:** "Whitespace" in this context refers to ASCII space (U+0020) and horizontal tab (U+0009). Such whitespace is not permitted in the following locations:
    * Between the opening `{` and an inline operator (e.g., `{ +` is invalid).
    * Between a structural keyword and its delimiter (e.g., `{move ~` or `{move :` are invalid).
    * Between the content-closing operator of an inline edit and an optional editor ID.
    * Between an editor ID and the final `}` in an inline edit.
    * Within a `TAG` identifier.
    If such forbidden whitespace exists, the sequence will not be recognized as valid EditML and will likely fall into 'Graceful Failure' or be treated as plain text. Other Unicode whitespace characters or newlines are also not permitted in these specific syntax-defining locations. Whitespace *within* user `text` or `block content` is preserved and does not require escaping unless it forms part of a delimiter sequence that needs escaping.
* Handle escape sequences by converting them to their literal character equivalents for content strings in the AST.

### 4.3. Abstract Syntax Tree (AST)
Key node types: `DocumentNode`, `TextNode`, `InlineEditNode` (type, content, optional ID, position, validity), `StructuralNode` (type, tag, block content for sources, position, validity). `block content` for source nodes should be stored by the parser as the clean, unescaped string.

### 4.4. Validation
* Ensure balanced, unescaped `{` and `}` for EditML blocks.
* Validate operator usage in `bbtext`.
* Validate tag naming (strictly alphanumeric as per 3.4.1 and 3.4.2).
* Enforce structural rules (unique source tags, move target multiplicity).
* Log issues (errors/warnings) with positions; mark AST nodes as invalid if needed, but attempt to produce a partial AST.

### 4.5. Graceful Failure (Unknown `{` Blocks)
If a block starts with an unescaped `{` but does not conform to defined EditML syntax:
1.  Identify its full extent by finding the first corresponding unescaped closing `}`. An unescaped closing `}` is one not immediately preceded by an odd number of backslashes `\`. The search for this `}` should correctly handle escaped `\}` sequences within the content of the unknown block.
2.  During this scan for the closing `}` of an unknown block, the parser does not attempt to interpret or validate any potential nested EditML structures or other EditML syntax within the unknown block's content; it only seeks the correctly matched, unescaped closing `}`.
3.  Treat the entire content from the opening `{` to this closing `}` (inclusive) as a single `TextNode`.
4.  Optionally log a warning.

## 5. Transformation & API

### 5.1. Transformation Profiles
* **`CleanView` Profile:** Additions applied, Deletions removed, Comments removed, Highlights become plain text. Structural edits applied; unresolved tags preserved as literal text.
* **`MarkupView` Profile:** All EditML markup preserved literally.
* **`HTMLPreview` Profile:** `bbtext` visually styled. `bbstructure` tags typically rendered as literal markup.

#### 5.1.1. Structural Edit Execution Order
When applying structural edits (e.g., in `CleanView`):
1.  **Resolution:** Identify all valid source-target pairs.
2.  **Order:** A recommended deterministic order:
    * All `copy` operations are performed first, based on the document order of their source tags. Content is copied to all valid targets.
    * Then, all `move` operations are performed, based on the document order of their source tags. Content is moved to its valid target.
    (This order prioritizes the stability of source content for copies before moves might eliminate that content).
3.  **Conflict Resolution:** If structural operations conflict (e.g., two `move` sources attempt to move overlapping content, a `copy` source overlaps a `move` source, or multiple `move` targets exist for the same tag as defined in 3.4.3), this specification considers it an error condition. Implementations *must* report such conflicts. The default and recommended behavior is that the transformation process for *all* structural edits in the document *must* be aborted to prevent a partially applied or inconsistent state and maintain data integrity. If an implementation offers an alternative strategy (e.g., applying only non-conflicting edits or using a specific, clearly documented precedence rule), this deviation from the recommended default behavior must be clearly documented and potentially user-configurable.

### 5.2. API Functions
* `ProcessDocument(inputText string) (*model.Document, error)`: Parses, validates, returns AST and issues.
* `TransformDocument(doc *model.Document, profile TransformationProfile) (string, error)`: Applies profile, returns string.

## 6. Rendering (Optional HTML Preview)
(As previously defined: `ins`, `del`, `span` with classes for `bbtext`).

## 7. Recommended Package Layout (Go)
(As previously defined).

---
*A formal grammar (e.g., EBNF/ABNF) is not included in this version but is acknowledged as a valuable future enhancement for achieving maximum parsing precision and interoperability across diverse implementations.*



