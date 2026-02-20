package lsp

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/github/gh-aw/pkg/logger"
	"github.com/github/gh-aw/pkg/parser"

	goyaml "go.yaml.in/yaml/v3"
)

var schemaLog = logger.New("lsp:schema")

// SchemaProvider extracts completion and hover info from the embedded JSON schema.
type SchemaProvider struct {
	schema map[string]any
}

// NewSchemaProvider creates a schema provider from the embedded main workflow schema.
func NewSchemaProvider() (*SchemaProvider, error) {
	raw := parser.GetMainWorkflowSchema()
	var schema map[string]any
	if err := json.Unmarshal([]byte(raw), &schema); err != nil {
		return nil, fmt.Errorf("parsing schema: %w", err)
	}
	return &SchemaProvider{schema: schema}, nil
}

// PropertyInfo holds extracted info about a schema property.
type PropertyInfo struct {
	Name        string
	Description string
	Type        string
	Enum        []string
	Default     string
	Deprecated  bool
	Required    bool
}

// TopLevelProperties returns info about all top-level frontmatter properties.
func (sp *SchemaProvider) TopLevelProperties() []PropertyInfo {
	return sp.propertiesAt(sp.schema)
}

// NestedProperties returns info about properties at a given YAML path (e.g., ["on"]).
func (sp *SchemaProvider) NestedProperties(path []string) []PropertyInfo {
	node := sp.resolveSchemaPath(path)
	if node == nil {
		return nil
	}
	return sp.propertiesAt(node)
}

// PropertyDescription returns hover information for the property at the given path.
func (sp *SchemaProvider) PropertyDescription(path []string) *PropertyInfo {
	if len(path) == 0 {
		return nil
	}

	// Navigate to the parent, then look up the property
	parent := sp.schema
	for i := 0; i < len(path)-1; i++ {
		parent = sp.resolveOneLevel(parent, path[i])
		if parent == nil {
			return nil
		}
	}

	key := path[len(path)-1]
	props := sp.getProperties(parent)
	if props == nil {
		return nil
	}

	propDef, ok := props[key].(map[string]any)
	if !ok {
		return nil
	}

	// Follow $ref if present
	propDef = sp.resolveRef(propDef)

	info := sp.extractPropertyInfo(key, propDef)
	return &info
}

// EnumValues returns enum values for a property at the given path.
func (sp *SchemaProvider) EnumValues(path []string) []string {
	if len(path) == 0 {
		return nil
	}

	// Navigate to the property
	node := sp.schema
	for i := 0; i < len(path)-1; i++ {
		node = sp.resolveOneLevel(node, path[i])
		if node == nil {
			return nil
		}
	}

	key := path[len(path)-1]
	props := sp.getProperties(node)
	if props == nil {
		return nil
	}

	propDef, ok := props[key].(map[string]any)
	if !ok {
		return nil
	}

	propDef = sp.resolveRef(propDef)

	return sp.extractEnums(propDef)
}

// resolveSchemaPath navigates the schema to find the object at the given path.
func (sp *SchemaProvider) resolveSchemaPath(path []string) map[string]any {
	node := sp.schema
	for _, key := range path {
		node = sp.resolveOneLevel(node, key)
		if node == nil {
			return nil
		}
	}
	return node
}

// resolveOneLevel navigates one key deeper in the schema.
func (sp *SchemaProvider) resolveOneLevel(node map[string]any, key string) map[string]any {
	props := sp.getProperties(node)
	if props == nil {
		return nil
	}

	propDef, ok := props[key].(map[string]any)
	if !ok {
		return nil
	}

	return sp.resolveRef(propDef)
}

// resolveRef follows a $ref if present in the schema node.
func (sp *SchemaProvider) resolveRef(node map[string]any) map[string]any {
	ref, ok := node["$ref"].(string)
	if !ok {
		return node
	}

	// Handle internal refs like "#/$defs/engine_config"
	if strings.HasPrefix(ref, "#/") {
		parts := strings.Split(strings.TrimPrefix(ref, "#/"), "/")
		resolved := sp.schema
		for _, part := range parts {
			next, ok := resolved[part].(map[string]any)
			if !ok {
				return node
			}
			resolved = next
		}
		return resolved
	}

	return node
}

// getProperties extracts the "properties" map from a schema node, handling oneOf/anyOf.
func (sp *SchemaProvider) getProperties(node map[string]any) map[string]any {
	if props, ok := node["properties"].(map[string]any); ok {
		return props
	}

	// Check oneOf/anyOf for object type with properties
	for _, key := range []string{"oneOf", "anyOf"} {
		if options, ok := node[key].([]any); ok {
			for _, opt := range options {
				optMap, ok := opt.(map[string]any)
				if !ok {
					continue
				}
				optMap = sp.resolveRef(optMap)
				if props, ok := optMap["properties"].(map[string]any); ok {
					return props
				}
			}
		}
	}

	return nil
}

// propertiesAt returns PropertyInfo for all properties defined in the given schema node.
func (sp *SchemaProvider) propertiesAt(node map[string]any) []PropertyInfo {
	props := sp.getProperties(node)
	if props == nil {
		return nil
	}

	// Build a set of required properties
	requiredSet := make(map[string]bool)
	if reqArr, ok := node["required"].([]any); ok {
		for _, r := range reqArr {
			if s, ok := r.(string); ok {
				requiredSet[s] = true
			}
		}
	}

	var result []PropertyInfo
	for key, val := range props {
		propDef, ok := val.(map[string]any)
		if !ok {
			continue
		}
		propDef = sp.resolveRef(propDef)
		info := sp.extractPropertyInfo(key, propDef)
		info.Required = requiredSet[key]
		result = append(result, info)
	}

	return result
}

// extractPropertyInfo extracts PropertyInfo from a property definition.
func (sp *SchemaProvider) extractPropertyInfo(key string, propDef map[string]any) PropertyInfo {
	info := PropertyInfo{Name: key}

	if desc, ok := propDef["description"].(string); ok {
		info.Description = desc
	}

	if t, ok := propDef["type"].(string); ok {
		info.Type = t
	}

	if def, ok := propDef["default"]; ok {
		info.Default = fmt.Sprintf("%v", def)
	}

	if dep, ok := propDef["deprecated"].(bool); ok {
		info.Deprecated = dep
	}

	// Check for description containing "DEPRECATED"
	if strings.Contains(strings.ToUpper(info.Description), "DEPRECATED") {
		info.Deprecated = true
	}

	info.Enum = sp.extractEnums(propDef)

	return info
}

// extractEnums extracts enum values from a property, including from oneOf options.
func (sp *SchemaProvider) extractEnums(propDef map[string]any) []string {
	if enumVals, ok := propDef["enum"].([]any); ok {
		var enums []string
		for _, v := range enumVals {
			enums = append(enums, fmt.Sprintf("%v", v))
		}
		return enums
	}

	// Check oneOf for string enums
	if options, ok := propDef["oneOf"].([]any); ok {
		for _, opt := range options {
			optMap, ok := opt.(map[string]any)
			if !ok {
				continue
			}
			optMap = sp.resolveRef(optMap)
			if t, ok := optMap["type"].(string); ok && t == "string" {
				if enumVals, ok := optMap["enum"].([]any); ok {
					var enums []string
					for _, v := range enumVals {
						enums = append(enums, fmt.Sprintf("%v", v))
					}
					return enums
				}
			}
		}
	}

	return nil
}

// YAMLPathAtPosition computes the YAML key path at a given position in the frontmatter.
// Returns the path (e.g., ["on", "issues", "types"]) and the current key at the cursor.
func YAMLPathAtPosition(yamlContent string, line int) (path []string, currentKey string) {
	if yamlContent == "" {
		return nil, ""
	}

	var root goyaml.Node
	if err := goyaml.Unmarshal([]byte(yamlContent), &root); err != nil {
		schemaLog.Printf("YAML parse error in path resolution: %v", err)
		return yamlPathFallback(yamlContent, line)
	}

	if root.Kind != goyaml.DocumentNode || len(root.Content) == 0 {
		return yamlPathFallback(yamlContent, line)
	}

	mapping := root.Content[0]
	if mapping.Kind != goyaml.MappingNode {
		return yamlPathFallback(yamlContent, line)
	}

	return findPathInMapping(mapping, line, nil)
}

// findPathInMapping recursively finds the YAML path at a given line within a mapping node.
func findPathInMapping(mapping *goyaml.Node, targetLine int, parentPath []string) ([]string, string) {
	if mapping.Kind != goyaml.MappingNode {
		return parentPath, ""
	}

	for i := 0; i+1 < len(mapping.Content); i += 2 {
		keyNode := mapping.Content[i]
		valNode := mapping.Content[i+1]

		keyName := keyNode.Value

		// yaml.v3 uses 1-based line numbers
		keyLine := keyNode.Line - 1

		if keyLine == targetLine {
			return parentPath, keyName
		}

		// Check if the target line falls within this value's range
		if valNode.Kind == goyaml.MappingNode {
			// Determine the end line of this mapping value
			valEndLine := nodeEndLine(valNode)
			if targetLine > keyLine && targetLine <= valEndLine {
				newPath := append(append([]string{}, parentPath...), keyName)
				return findPathInMapping(valNode, targetLine, newPath)
			}
		} else if valNode.Kind == goyaml.SequenceNode {
			valEndLine := nodeEndLine(valNode)
			if targetLine > keyLine && targetLine <= valEndLine {
				return append(append([]string{}, parentPath...), keyName), ""
			}
		}
	}

	return parentPath, ""
}

// nodeEndLine returns the last line number (0-based) of a YAML node and its children.
func nodeEndLine(node *goyaml.Node) int {
	maxLine := node.Line - 1
	for _, child := range node.Content {
		childEnd := nodeEndLine(child)
		if childEnd > maxLine {
			maxLine = childEnd
		}
	}
	return maxLine
}

// yamlPathFallback uses indentation-based heuristic when AST is unavailable (e.g., incomplete YAML).
func yamlPathFallback(yamlContent string, targetLine int) ([]string, string) {
	lines := strings.Split(yamlContent, "\n")
	if targetLine < 0 || targetLine >= len(lines) {
		return nil, ""
	}

	currentLine := lines[targetLine]
	trimmed := strings.TrimSpace(currentLine)

	// Determine current key
	currentKey := ""
	if idx := strings.Index(trimmed, ":"); idx > 0 {
		currentKey = strings.TrimSpace(trimmed[:idx])
	}

	// Build path by walking up lines with decreasing indentation
	currentIndent := len(currentLine) - len(strings.TrimLeft(currentLine, " "))
	var path []string

	for i := targetLine - 1; i >= 0; i-- {
		line := lines[i]
		if strings.TrimSpace(line) == "" {
			continue
		}
		indent := len(line) - len(strings.TrimLeft(line, " "))
		if indent < currentIndent {
			t := strings.TrimSpace(line)
			if idx := strings.Index(t, ":"); idx > 0 {
				path = append([]string{strings.TrimSpace(t[:idx])}, path...)
			}
			currentIndent = indent
		}
	}

	return path, currentKey
}
