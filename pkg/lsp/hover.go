package lsp

import (
	"fmt"
	"strings"

	"github.com/github/gh-aw/pkg/logger"
)

var hoverLog = logger.New("lsp:hover")

// HandleHover computes hover information for a position in a document.
func HandleHover(snap *DocumentSnapshot, pos Position, sp *SchemaProvider) *Hover {
	if snap == nil || sp == nil {
		return nil
	}

	// Only provide hover inside frontmatter
	if !snap.PositionInFrontmatter(pos) {
		return nil
	}

	// Compute the YAML path at the cursor position
	// The line in the YAML content is relative to the frontmatter start
	yamlLine := pos.Line - snap.FrontmatterStartLine - 1
	path, currentKey := YAMLPathAtPosition(snap.FrontmatterYAML, yamlLine)

	hoverLog.Printf("Hover at line %d: path=%v, key=%s", pos.Line, path, currentKey)

	if currentKey == "" {
		return nil
	}

	// Look up the property in the schema
	fullPath := append(path, currentKey)
	info := sp.PropertyDescription(fullPath)
	if info == nil {
		return nil
	}

	return &Hover{
		Contents: MarkupContent{
			Kind:  "markdown",
			Value: formatHoverContent(info),
		},
	}
}

// formatHoverContent formats a PropertyInfo into markdown for hover display.
func formatHoverContent(info *PropertyInfo) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("### `%s`\n\n", info.Name))

	if info.Deprecated {
		sb.WriteString("⚠️ **Deprecated**\n\n")
	}

	if info.Description != "" {
		sb.WriteString(info.Description)
		sb.WriteString("\n\n")
	}

	if info.Type != "" {
		sb.WriteString(fmt.Sprintf("**Type:** `%s`\n\n", info.Type))
	}

	if info.Default != "" {
		sb.WriteString(fmt.Sprintf("**Default:** `%s`\n\n", info.Default))
	}

	if info.Required {
		sb.WriteString("**Required**\n\n")
	}

	if len(info.Enum) > 0 {
		sb.WriteString("**Allowed values:** ")
		quoted := make([]string, len(info.Enum))
		for i, e := range info.Enum {
			quoted[i] = "`" + e + "`"
		}
		sb.WriteString(strings.Join(quoted, ", "))
		sb.WriteString("\n")
	}

	return strings.TrimSpace(sb.String())
}
