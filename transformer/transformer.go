// transformer/transformer.go
// package transformer provides functionality to transform an EditML AST into output strings.
package transformer

import (
	"fmt"
	"strings"

	"github.com/verkaro/editml-go/model"
	"github.com/verkaro/editml-go/parser"
)

// TransformToCleanView is the internal function that takes a slice of nodes (AST)
// and applies transformations to produce a "Clean View" string.
// It also returns any critical errors encountered during transformation.
// This function is unexported and will be called by the public editml.TransformCleanView().
func TransformToCleanView(nodes []model.Node) (string, error) { // Renamed from transformToCleanView
	// This implementation is adapted from the Backburner POC's transformer.
	// It focuses on "CleanView": additions applied, deletions/comments removed,
	// highlights as plain text, structural edits resolved.

	// --- Step 1: Pre-scan to collect structural operations and detect immediate conflicts ---
	// This logic is similar to the Backburner POC.
	type sourceDetail struct {
		node             model.StructuralSourceNode
		transformedBlock string // Pre-transformed content of the source block
		isUsedAsMove     bool   // True if this source is part of a successful move operation
	}
	allSources := make(map[string]*sourceDetail) // Tag -> *sourceDetail

	type targetDetail struct {
		node model.StructuralTargetNode
	}
	allTargets := make(map[string][]targetDetail) // Tag -> []targetDetail

	moveTargetCounts := make(map[string]int) // For move conflict detection (multiple move targets)

	// First pass: Collect sources, targets, and pre-transform source BlockContent.
	// Also, validate structural rules that can be checked at this stage.
	for _, node := range nodes {
		if srcNode, ok := node.(model.StructuralSourceNode); ok {
			// Check for duplicate source tags (Spec 3.4.3)
			if _, exists := allSources[srcNode.Tag]; exists {
				// For MVP, this is a critical error.
				// Future: Could be an editml.Issue with more detail.
				return "", fmt.Errorf("structural conflict: duplicate source tag %q", srcNode.Tag)
			}

			// Pre-transform the block content of the source node.
			// The BlockContent itself can contain inline EditML.
			// MVP: We re-parse and transform the BlockContent string here.
			// Future: If BlockContent is []model.Node in AST, this re-parsing isn't needed.
			// Note: parser.ParseEditMLToNodes is already exported.
			subParserNodes, parseErr := parser.ParseEditMLToNodes(srcNode.BlockContent)
			transformedBlock := ""
			if parseErr != nil {
				// If BlockContent parsing fails, this is a problem for the structural operation.
				// For CleanView, an error in content might mean the structural op is "broken".
				// For MVP, we'll represent this as an error in the transformed block.
				// Future: This should generate an editml.Issue.
				transformedBlock = fmt.Sprintf("{%s~%s (ERROR_PARSING_CONTENT)~%s}", srcNode.Operation, srcNode.BlockContent, srcNode.Tag)
			} else {
				// Recursively call this transformer for the sub-nodes.
				// IMPORTANT: This recursive call assumes that the sub-nodes (from BlockContent)
				// do NOT contain further structural tags that could interact with the outer document's
				// structural tags. Spec 3.4.3 states bbstructure cannot be nested, so the parser
				// should ideally prevent this. If nested structural tags were parsed into subParserNodes,
				// this could lead to complex behavior or infinite loops if not handled carefully.
				// For MVP, we rely on the parser not producing nested structural tags within BlockContent's AST.
				blockStr, transformErr := TransformToCleanView(subParserNodes) // Recursive call, now capitalized
				if transformErr != nil {
					transformedBlock = fmt.Sprintf("{%s~%s (ERROR_TRANSFORMING_CONTENT)~%s}", srcNode.Operation, srcNode.BlockContent, srcNode.Tag)
				} else {
					transformedBlock = blockStr
				}
			}
			allSources[srcNode.Tag] = &sourceDetail{node: srcNode, transformedBlock: transformedBlock}

		} else if targetNode, ok := node.(model.StructuralTargetNode); ok {
			allTargets[targetNode.Tag] = append(allTargets[targetNode.Tag], targetDetail{node: targetNode})
			if targetNode.Operation == model.OperationMove {
				moveTargetCounts[targetNode.Tag]++
				if moveTargetCounts[targetNode.Tag] > 1 {
					// Spec 3.4.3: Multiple move targets for the same tag is an error.
					return "", fmt.Errorf("structural conflict: multiple move targets for tag %q", targetNode.Tag)
				}
			}
		}
	}

	// Mark move sources that have a valid target.
	// These sources won't render their content at their original location.
	for tag, srcDetail := range allSources {
		if srcDetail.node.Operation == model.OperationMove {
			if targets, hasTargets := allTargets[tag]; hasTargets && len(targets) > 0 {
				// A move source is considered "used" if there's exactly one corresponding move target.
				// The moveTargetCounts check above already ensures no more than one move target.
				isMoveTargetForThisSource := false
				for _, t := range targets {
					if t.node.Operation == model.OperationMove {
						isMoveTargetForThisSource = true
						break
					}
				}
				if isMoveTargetForThisSource && moveTargetCounts[tag] == 1 {
					srcDetail.isUsedAsMove = true
				}
			}
		}
	}

	// --- Step 2: Build the output string by applying transformations ---
	var sb strings.Builder

	for _, node := range nodes {
		switch n := node.(type) {
		case model.TextNode:
			sb.WriteString(n.Text)
		case model.InlineEditNode:
			switch n.EditType {
			case model.EditTypeAddition:
				sb.WriteString(n.Content) // Apply addition
			case model.EditTypeDeletion:
				// Omitted in CleanView
			case model.EditTypeComment:
				// Omitted in CleanView
			case model.EditTypeHighlight:
				sb.WriteString(n.Content) // Highlight becomes plain text in CleanView
			}
		case model.StructuralSourceNode:
			srcDetail, sourceExists := allSources[n.Tag]
			if !sourceExists { // Should not happen if collected properly in Step 1
				// This indicates an internal inconsistency.
				// For MVP, render a placeholder indicating the error.
				// Future: This should be an internal error, potentially an editml.Issue.
				sb.WriteString(fmt.Sprintf("{%s~%s~%s (ERROR_SOURCE_NOT_FOUND_IN_MAP)}", n.Operation, n.BlockContent, n.Tag))
				continue
			}

			if n.Operation == model.OperationMove {
				if !srcDetail.isUsedAsMove {
					// Unresolved move source (no valid single move target found for this move operation)
					// or if the source's block content had errors during its transformation.
					// Spec 5.1.1: "unresolved tags preserved as literal text."
					// If transformedBlock contains an error message, use that; otherwise, render literally.
					if strings.Contains(srcDetail.transformedBlock, "(ERROR_") {
						sb.WriteString(srcDetail.transformedBlock)
					} else {
						sb.WriteString(fmt.Sprintf("{%s~%s~%s}", n.Operation, n.BlockContent, n.Tag))
					}
				}
				// If it isUsedAsMove, content is rendered by the target node, so do nothing here.
			} else if n.Operation == model.OperationCopy {
				// For copy, the source's (transformed) content appears at its original location
				// if it has valid targets. If no targets, it's an unresolved copy source.
				// Spec 5.1.1: "unresolved tags preserved as literal text."
				targetsForThisCopy, hasTargets := allTargets[n.Tag]
				isValidCopyOperation := false
				if hasTargets {
					for _, t := range targetsForThisCopy {
						if t.node.Operation == model.OperationCopy { // Ensure target is also a copy op
							isValidCopyOperation = true
							break
						}
					}
				}

				if !isValidCopyOperation || strings.Contains(srcDetail.transformedBlock, "(ERROR_") {
					// No valid copy targets, or error in block content: render copy source tag literally.
					if strings.Contains(srcDetail.transformedBlock, "(ERROR_") {
						sb.WriteString(srcDetail.transformedBlock)
					} else {
						sb.WriteString(fmt.Sprintf("{%s~%s~%s}", n.Operation, n.BlockContent, n.Tag))
					}
				} else {
					// Has valid copy targets and no error in block: render its transformed content at original location.
					sb.WriteString(srcDetail.transformedBlock)
				}
			}

		case model.StructuralTargetNode:
			srcDetail, sourceExists := allSources[n.Tag]
			if !sourceExists {
				// Unresolved target (no source defined for this tag).
				// Spec 5.1.1: "unresolved tags preserved as literal text."
				sb.WriteString(fmt.Sprintf("{%s:%s}", n.Operation, n.Tag))
				continue
			}

			// Check if the source and target operations match (e.g., move target for move source).
			// Spec 3.4.3: "No Dual Operation Type for a Tag" implies target op should match source op.
			if n.Operation != srcDetail.node.Operation {
				// This is a structural conflict.
				// Future: This should be an editml.Issue.
				// For MVP, render a placeholder.
				sb.WriteString(fmt.Sprintf("{%s:%s (ERROR_OPERATION_MISMATCH_WITH_SOURCE %s)}", n.Operation, n.Tag, srcDetail.node.Operation))
				continue
			}

			// If the source block had transformation errors, reflect that at the target.
			if strings.Contains(srcDetail.transformedBlock, "(ERROR_") {
				sb.WriteString(srcDetail.transformedBlock)
				continue
			}

			if n.Operation == model.OperationMove {
				// Content was marked by isUsedAsMove on the source.
				// The actual rendering of moved content happens here at the target.
				// The pre-scan already confirmed only one valid move target if isUsedAsMove is true.
				if srcDetail.isUsedAsMove { // Double check if this move target corresponds to a used move source
					sb.WriteString(srcDetail.transformedBlock)
				} else {
					// This case implies a move target whose corresponding move source was not "used"
					// (e.g. source was invalid, or this target is somehow orphaned despite matching tag).
					// Render as unresolved.
					sb.WriteString(fmt.Sprintf("{%s:%s}", n.Operation, n.Tag))
				}
			} else if n.Operation == model.OperationCopy {
				// For copy, write the pre-transformed content at each valid copy target.
				sb.WriteString(srcDetail.transformedBlock)
			}
		}
	}
	return sb.String(), nil
}
