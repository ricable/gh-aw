package lsp

import (
	"fmt"
	"sort"
	"strings"

	"github.com/github/gh-aw/pkg/logger"
)

var completionLog = logger.New("lsp:completion")

// HandleCompletion computes completion items for a position in a document.
func HandleCompletion(snap *DocumentSnapshot, pos Position, sp *SchemaProvider) *CompletionList {
	if snap == nil || sp == nil {
		return &CompletionList{}
	}

	// If no frontmatter, suggest the skeleton snippet
	if !snap.HasFrontmatter {
		return &CompletionList{
			Items: snippetCompletions(),
		}
	}

	// Only provide completions inside frontmatter
	if !snap.PositionInFrontmatter(pos) {
		return &CompletionList{}
	}

	// Compute the YAML path at the cursor position
	yamlLine := pos.Line - snap.FrontmatterStartLine - 1
	path, currentKey := YAMLPathAtPosition(snap.FrontmatterYAML, yamlLine)

	completionLog.Printf("Completion at line %d: path=%v, key=%s", pos.Line, path, currentKey)

	var items []CompletionItem

	if len(path) == 0 && currentKey == "" {
		// At top level with no context — suggest top-level keys
		items = topLevelKeyCompletions(sp)
	} else if len(path) == 0 && currentKey != "" {
		// Typing a top-level key — filter top-level keys
		items = filterCompletions(topLevelKeyCompletions(sp), currentKey)
	} else {
		// Nested context — first check if there are enum values for the value
		if currentKey != "" {
			enums := sp.EnumValues(append(path, currentKey))
			if len(enums) > 0 {
				items = enumCompletions(enums)
			}
		}

		// Also suggest nested keys at this path
		if len(items) == 0 {
			props := sp.NestedProperties(path)
			if len(props) > 0 {
				items = propertyCompletions(props)
			}
		}
	}

	// Always include snippets when at top level
	if len(path) == 0 {
		items = append(items, snippetCompletions()...)
	}

	return &CompletionList{
		IsIncomplete: false,
		Items:        items,
	}
}

// topLevelKeyCompletions returns completion items for all top-level frontmatter keys.
func topLevelKeyCompletions(sp *SchemaProvider) []CompletionItem {
	props := sp.TopLevelProperties()
	return propertyCompletions(props)
}

// propertyCompletions converts PropertyInfo items into CompletionItem items.
func propertyCompletions(props []PropertyInfo) []CompletionItem {
	sort.Slice(props, func(i, j int) bool {
		// Required first, then alphabetical
		if props[i].Required != props[j].Required {
			return props[i].Required
		}
		return props[i].Name < props[j].Name
	})

	items := make([]CompletionItem, 0, len(props))
	for i, p := range props {
		item := CompletionItem{
			Label:      p.Name,
			Kind:       CompletionItemKindProperty,
			Detail:     p.Type,
			Deprecated: p.Deprecated,
		}

		if p.Description != "" {
			item.Documentation = &MarkupContent{
				Kind:  "markdown",
				Value: p.Description,
			}
		}

		if p.Required {
			item.Detail = p.Type + " (required)"
		}

		// Insert text with colon and space
		item.InsertText = p.Name + ": "
		item.InsertTextFormat = InsertTextFormatPlainText

		// Sort required items first using prefix
		if p.Required {
			item.SortText = "0_" + p.Name
		} else {
			item.SortText = "1_" + padIndex(i)
		}

		items = append(items, item)
	}

	return items
}

// enumCompletions returns completion items for enum values.
func enumCompletions(values []string) []CompletionItem {
	items := make([]CompletionItem, 0, len(values))
	for _, v := range values {
		items = append(items, CompletionItem{
			Label:            v,
			Kind:             CompletionItemKindEnum,
			InsertText:       v,
			InsertTextFormat: InsertTextFormatPlainText,
		})
	}
	return items
}

// snippetCompletions returns pre-built snippet completions for common workflow patterns.
func snippetCompletions() []CompletionItem {
	return []CompletionItem{
		{
			Label:  "aw: Minimal workflow",
			Kind:   CompletionItemKindSnippet,
			Detail: "New agentic workflow skeleton",
			Documentation: &MarkupContent{
				Kind:  "markdown",
				Value: "Creates a minimal agentic workflow with issue trigger, Copilot engine, and safe outputs.",
			},
			InsertText: `---
on:
  issues:
    types: [opened]
engine: copilot
permissions: read-all
safe-outputs:
  add-comment:
---
# $1

$0`,
			InsertTextFormat: InsertTextFormatSnippet,
			SortText:         "2_snippet_minimal",
		},
		{
			Label:  "aw: Slash command workflow",
			Kind:   CompletionItemKindSnippet,
			Detail: "Workflow triggered by slash command",
			Documentation: &MarkupContent{
				Kind:  "markdown",
				Value: "Creates a workflow triggered by a slash command in issue or PR comments.",
			},
			InsertText: `---
on:
  issue_comment:
    types: [created]
  slash_command:
    name: $1
engine: copilot
permissions: read-all
safe-outputs:
  add-comment:
---
# $2

$0`,
			InsertTextFormat: InsertTextFormatSnippet,
			SortText:         "2_snippet_slash",
		},
		{
			Label:  "aw: Workflow with imports",
			Kind:   CompletionItemKindSnippet,
			Detail: "Workflow with imported shared components",
			Documentation: &MarkupContent{
				Kind:  "markdown",
				Value: "Creates a workflow that imports shared workflow components.",
			},
			InsertText: `---
on:
  issues:
    types: [opened]
engine: copilot
imports:
  - $1
permissions: read-all
safe-outputs:
  add-comment:
---
# $2

$0`,
			InsertTextFormat: InsertTextFormatSnippet,
			SortText:         "2_snippet_imports",
		},
	}
}

// filterCompletions filters completion items by prefix.
func filterCompletions(items []CompletionItem, prefix string) []CompletionItem {
	if prefix == "" {
		return items
	}
	prefix = strings.ToLower(prefix)
	var filtered []CompletionItem
	for _, item := range items {
		if strings.HasPrefix(strings.ToLower(item.Label), prefix) {
			filtered = append(filtered, item)
		}
	}
	return filtered
}

func padIndex(i int) string {
	return fmt.Sprintf("%04d", i)
}
