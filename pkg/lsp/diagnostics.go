package lsp

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/github/gh-aw/pkg/logger"
	"github.com/github/gh-aw/pkg/parser"

	goyaml "go.yaml.in/yaml/v3"
)

var diagLog = logger.New("lsp:diagnostics")

// ComputeDiagnostics produces diagnostics for a document snapshot.
func ComputeDiagnostics(snap *DocumentSnapshot) []Diagnostic {
	var diags []Diagnostic

	// 1. Check for frontmatter presence
	if !snap.HasFrontmatter {
		if len(snap.Lines) > 0 && strings.TrimSpace(snap.Lines[0]) != "---" {
			diags = append(diags, Diagnostic{
				Range:    lineRange(0),
				Severity: SeverityWarning,
				Source:   "gh-aw",
				Message:  "Workflow file is missing frontmatter (--- delimiters). Add frontmatter with at least an 'on' trigger.",
			})
		}
		return diags
	}

	// 2. Check for unclosed frontmatter (should not happen if HasFrontmatter is true, but defensively check)
	// Also check for multiple frontmatter blocks
	fmCount := 0
	for i, line := range snap.Lines {
		if strings.TrimSpace(line) == "---" {
			fmCount++
			if fmCount > 2 {
				diags = append(diags, Diagnostic{
					Range:    lineRange(i),
					Severity: SeverityWarning,
					Source:   "gh-aw",
					Message:  "Multiple frontmatter delimiters detected. Only the first frontmatter block is used.",
				})
			}
		}
	}

	// 3. YAML parse errors
	yamlDiags := checkYAMLSyntax(snap)
	diags = append(diags, yamlDiags...)

	// If YAML doesn't parse, skip schema validation
	if len(yamlDiags) > 0 {
		return diags
	}

	// 4. Schema validation
	schemaDiags := checkSchemaValidation(snap)
	diags = append(diags, schemaDiags...)

	return diags
}

// checkYAMLSyntax validates YAML syntax and returns diagnostics for parse errors.
func checkYAMLSyntax(snap *DocumentSnapshot) []Diagnostic {
	if snap.FrontmatterYAML == "" {
		return nil
	}

	var node goyaml.Node
	err := goyaml.Unmarshal([]byte(snap.FrontmatterYAML), &node)
	if err == nil {
		return nil
	}

	diagLog.Printf("YAML parse error: %v", err)

	// Try to extract line info from the error
	diag := Diagnostic{
		Range:    lineRange(snap.FrontmatterStartLine + 1), // Default to first content line
		Severity: SeverityError,
		Source:   "gh-aw",
		Message:  fmt.Sprintf("YAML syntax error: %s", err.Error()),
	}

	// go-yaml errors often contain "line N:" info
	errMsg := err.Error()
	if lineNum := extractYAMLErrorLine(errMsg); lineNum > 0 {
		// YAML error lines are 1-based and relative to the YAML content
		adjustedLine := snap.FrontmatterStartLine + lineNum
		if adjustedLine < len(snap.Lines) {
			diag.Range = lineRange(adjustedLine)
		}
	}

	return []Diagnostic{diag}
}

// extractYAMLErrorLine extracts the line number from a YAML error message.
func extractYAMLErrorLine(errMsg string) int {
	// Look for "line N" pattern
	idx := strings.Index(errMsg, "line ")
	if idx < 0 {
		return 0
	}
	numStr := ""
	for i := idx + 5; i < len(errMsg); i++ {
		if errMsg[i] >= '0' && errMsg[i] <= '9' {
			numStr += string(errMsg[i])
		} else {
			break
		}
	}
	if numStr == "" {
		return 0
	}
	var n int
	if _, err := fmt.Sscanf(numStr, "%d", &n); err != nil {
		return 0
	}
	return n
}

// checkSchemaValidation validates frontmatter against the JSON schema and returns diagnostics.
func checkSchemaValidation(snap *DocumentSnapshot) []Diagnostic {
	var frontmatter map[string]any

	if snap.FrontmatterYAML != "" {
		// Parse the YAML into a map
		if err := goyaml.Unmarshal([]byte(snap.FrontmatterYAML), &frontmatter); err != nil {
			return nil // YAML errors already reported by checkYAMLSyntax
		}
	}

	if frontmatter == nil {
		frontmatter = make(map[string]any)
	}

	// Use the same validation the compiler uses: normalize through JSON
	frontmatterJSON, err := json.Marshal(frontmatter)
	if err != nil {
		return nil
	}

	var normalized any
	if err := json.Unmarshal(frontmatterJSON, &normalized); err != nil {
		return nil
	}

	// Validate using the parser package's schema validation
	schemaErr := parser.ValidateMainWorkflowFrontmatterWithSchema(frontmatter)
	if schemaErr == nil {
		return nil
	}

	diagLog.Printf("Schema validation error: %v", schemaErr)

	// Convert the error to diagnostics
	return schemaErrorToDiagnostics(snap, schemaErr)
}

// schemaErrorToDiagnostics converts a schema validation error into diagnostics.
func schemaErrorToDiagnostics(snap *DocumentSnapshot, err error) []Diagnostic {
	errMsg := err.Error()

	// Default: attach to the frontmatter start line
	defaultRange := lineRange(snap.FrontmatterStartLine + 1)

	// Try to extract field paths from the error message to improve location
	// Typical schema errors mention paths like "on", "engine", etc.
	diag := Diagnostic{
		Range:    defaultRange,
		Severity: SeverityError,
		Source:   "gh-aw",
		Message:  cleanSchemaErrorMessage(errMsg),
	}

	// Try to locate the relevant line in frontmatter
	if snap.FrontmatterYAML != "" {
		if line := findRelevantLine(snap, errMsg); line >= 0 {
			diag.Range = lineRange(line)
		}
	}

	return []Diagnostic{diag}
}

// findRelevantLine tries to find the line in the document that a schema error refers to.
func findRelevantLine(snap *DocumentSnapshot, errMsg string) int {
	// Extract field names mentioned in schema errors
	// Common patterns: "missing properties: 'on'", "'engine' ...", "/properties/on"
	topLevelKeys := []string{
		"on", "engine", "tools", "safe-outputs", "safe-inputs", "permissions",
		"imports", "network", "sandbox", "name", "description",
	}

	for _, key := range topLevelKeys {
		if !strings.Contains(errMsg, "'"+key+"'") && !strings.Contains(errMsg, "/"+key) {
			continue
		}
		// Find this key in the frontmatter lines
		for i := snap.FrontmatterStartLine + 1; i < snap.FrontmatterEndLine; i++ {
			if i < len(snap.Lines) {
				trimmed := strings.TrimSpace(snap.Lines[i])
				if strings.HasPrefix(trimmed, key+":") || strings.HasPrefix(trimmed, key+" :") {
					return i
				}
			}
		}
	}

	return -1
}

// cleanSchemaErrorMessage cleans up verbose schema error messages.
func cleanSchemaErrorMessage(msg string) string {
	// Remove file path prefixes from console-formatted errors
	if idx := strings.Index(msg, "error:"); idx >= 0 {
		cleaned := strings.TrimSpace(msg[idx+6:])
		if cleaned != "" {
			return cleaned
		}
	}

	// Remove jsonschema prefix noise
	msg = strings.TrimPrefix(msg, "jsonschema validation failed with ")

	// Remove internal schema URL references
	if strings.Contains(msg, "http://contoso.com/") {
		// Extract the meaningful part: "missing property 'on'"
		if idx := strings.Index(msg, "missing property"); idx >= 0 {
			msg = strings.TrimSpace(msg[idx:])
		} else if idx := strings.Index(msg, "additional properties"); idx >= 0 {
			msg = strings.TrimSpace(msg[idx:])
		} else {
			// Strip the URL prefix line
			lines := strings.Split(msg, "\n")
			var cleaned []string
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if line == "" || strings.HasPrefix(line, "'http://contoso.com/") {
					continue
				}
				// Remove leading "- at '':" patterns
				if strings.HasPrefix(line, "- at") {
					if colonIdx := strings.Index(line, ":"); colonIdx >= 0 {
						line = strings.TrimSpace(line[colonIdx+1:])
					}
				}
				if line != "" {
					cleaned = append(cleaned, line)
				}
			}
			if len(cleaned) > 0 {
				msg = strings.Join(cleaned, "; ")
			}
		}
	}

	// Truncate very long messages
	if len(msg) > 500 {
		msg = msg[:497] + "..."
	}

	return msg
}

// lineRange creates a Range covering a full line.
func lineRange(line int) Range {
	return Range{
		Start: Position{Line: line, Character: 0},
		End:   Position{Line: line, Character: 1000},
	}
}
