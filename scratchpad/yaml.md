---
description: Why goccy/go-yaml was chosen and how to use it effectively in gh-aw
---

# YAML Library Choice and Usage Patterns

This document explains why `goccy/go-yaml` was chosen for GitHub Agentic Workflows and documents best practices for YAML handling.

## Table of Contents

- [Why goccy/go-yaml](#why-goccygo-yaml)
- [Library Comparison](#library-comparison)
- [Best Practices](#best-practices)
- [Usage Examples](#usage-examples)
- [Advanced Features](#advanced-features)
- [References](#references)

---

## Why goccy/go-yaml

GitHub Agentic Workflows uses **[goccy/go-yaml](https://github.com/goccy/go-yaml) v1.19.1** as its primary YAML library. This choice was made after careful evaluation of available Go YAML parsers.

### Key Advantages

1. **Superior YAML Test Suite Coverage** - 60+ more passing tests than go-yaml/yaml
2. **Better Error Messages** - Context-rich errors with source code snippets
3. **Active Maintenance** - Regular updates and bug fixes (commits as recent as 2026-02-16)
4. **Native Go Implementation** - Pure Go, not a C port (better for cross-compilation)
5. **YAML 1.2 Compliance** - Correct handling of `on:` triggers and boolean keywords
6. **Advanced Features** - YAML Path, AST manipulation, custom marshalers

### YAML 1.2 Compliance Benefits

YAML 1.2 compliance is critical for GitHub Actions compatibility. Unlike YAML 1.1 parsers (like PyYAML), `goccy/go-yaml` correctly treats `on`, `off`, `yes`, `no` as strings rather than booleans.

**Why this matters:**
- GitHub Actions uses `on:` as a trigger key (must be a string, not boolean)
- YAML 1.1 parsers would convert `on:` to `true:`, breaking workflow validation
- YAML 1.2 is the modern standard (released 2009)

See [yaml-version-gotchas.md](yaml-version-gotchas.md) for detailed comparison.

---

## Library Comparison

### goccy/go-yaml vs gopkg.in/yaml.v3

| Feature | goccy/go-yaml | gopkg.in/yaml.v3 |
|---------|---------------|------------------|
| **YAML Version** | 1.2 | 1.1 |
| **Error Messages** | Rich context with code snippets | Basic error messages |
| **Test Suite Passes** | 300+ | 240+ |
| **Maintenance** | Active (2026) | Moderate |
| **Implementation** | Pure Go | Pure Go |
| **MapSlice Support** | ✅ Yes | ✅ Yes |
| **Custom Marshalers** | ✅ Yes | ✅ Yes |
| **YAML Path** | ✅ Yes | ❌ No |
| **AST Access** | ✅ Yes | Limited |
| **Stream Processing** | ✅ Yes | ✅ Yes |

### When to Use Each Library

**Use goccy/go-yaml for:**
- ✅ Workflow frontmatter parsing
- ✅ GitHub Actions YAML generation
- ✅ Any YAML 1.2 sensitive operations
- ✅ User-facing YAML validation with helpful errors

**Use gopkg.in/yaml.v3 for:**
- Campaign specs (simple configuration)
- Workflow statistics (internal data)
- Simple marshaling where YAML 1.1/1.2 differences don't matter

**Note on Import Path:** Use the canonical `go.yaml.in/yaml/v3` path instead of the deprecated `gopkg.in/yaml.v3` for better supply chain security.

---

## Best Practices

The gh-aw codebase demonstrates excellent YAML handling patterns. These practices are already implemented throughout the project.

### 1. Deterministic Field Ordering

GitHub Actions workflows benefit from consistent field ordering for readability and maintainability.

**Problem:** Go maps have random iteration order, causing non-deterministic YAML output.

**Solution:** Use `yaml.MapSlice` to preserve insertion order.

```go
// ✅ CORRECT - Deterministic ordering using MapSlice
orderedData := yaml.MapSlice{
    {Key: "name", Value: "My Workflow"},
    {Key: "on", Value: triggers},
    {Key: "permissions", Value: perms},
    {Key: "jobs", Value: jobs},
}
```

### 2. Standard Marshal Options

Use consistent formatting options across the codebase for professional YAML output.

```go
// pkg/workflow/yaml_options.go
var DefaultMarshalOptions = []yaml.EncodeOption{
    yaml.Indent(2),                        // 2-space indentation (GitHub Actions standard)
    yaml.UseLiteralStyleIfMultiline(true), // Use | for multiline strings
}

// Usage
yamlBytes, err := yaml.MarshalWithOptions(data, DefaultMarshalOptions...)
```

**Benefits:**
- Consistent 2-space indentation across all workflows
- Literal block scalars (`|`) for readable multiline strings
- Matches GitHub Actions documentation style

### 3. Helper Abstractions

The workflow compiler provides clean abstractions for common YAML operations.

#### MarshalWithFieldOrder

Marshal maps with priority fields appearing first:

```go
// pkg/workflow/yaml.go
data := map[string]any{
    "jobs": map[string]any{...},
    "name": "My Workflow",
    "on": map[string]any{...},
    "permissions": map[string]any{...},
}

// Priority fields appear first, others alphabetically
yamlBytes, err := MarshalWithFieldOrder(data, []string{
    "name", "on", "permissions", "jobs",
})

// Result:
// name: My Workflow
// on: ...
// permissions: ...
// jobs: ...
```

#### OrderMapFields

Convert maps to ordered MapSlice structures:

```go
// Order job step fields conventionally
step := map[string]any{
    "env": map[string]string{"FOO": "bar"},
    "name": "Build project",
    "run": "make build",
}

orderedStep := OrderMapFields(step, []string{
    "name", "id", "if", "uses", "run", "with", "env",
})
// Result: name, run, env (in that order)
```

### 4. Post-Processing Utilities

Clean up YAML output to match GitHub conventions.

#### UnquoteYAMLKey

Remove unnecessary quotes from YAML keywords:

```go
// The marshaler quotes "on" to avoid YAML 1.1 boolean ambiguity
yamlStr := `"on":
  push:
    branches:
      - main`

// Remove quotes for cleaner output
cleaned := UnquoteYAMLKey(yamlStr, "on")
// Result:
// on:
//   push:
//     branches:
//       - main
```

#### CleanYAMLNullValues

Simplify null value representation:

```go
yamlStr := `on:
  workflow_dispatch: null
  schedule:
    - cron: '0 0 * * *'`

cleaned := CleanYAMLNullValues(yamlStr)
// Result:
// on:
//   workflow_dispatch:
//   schedule:
//     - cron: '0 0 * * *'
```

### 5. Error Handling

`goccy/go-yaml` provides rich error context:

```go
var data map[string]any
err := yaml.Unmarshal([]byte(content), &data)
if err != nil {
    // Error includes:
    // - Line and column numbers
    // - Source code snippet
    // - Contextual explanation
    return fmt.Errorf("failed to parse YAML: %w", err)
}
```

**Example error output:**
```
[3:5] unknown field "invalidsyntax"
   1 | name: Test
   2 | on:
>  3 |   invalidsyntax
         ^
   4 |     types: [opened]
```

---

## Usage Examples

### Example 1: Workflow Compilation

How gh-aw compiles frontmatter to GitHub Actions YAML:

```go
// pkg/workflow/compiler.go
func CompileWorkflow(frontmatter map[string]any) ([]byte, error) {
    // Step 1: Order top-level fields
    orderedWorkflow := OrderMapFields(frontmatter, []string{
        "name",
        "on",
        "permissions",
        "env",
        "defaults",
        "concurrency",
        "jobs",
    })
    
    // Step 2: Marshal with standard options
    yamlBytes, err := yaml.MarshalWithOptions(
        orderedWorkflow,
        DefaultMarshalOptions...,
    )
    if err != nil {
        return nil, fmt.Errorf("failed to marshal workflow: %w", err)
    }
    
    // Step 3: Post-process for GitHub conventions
    yamlStr := string(yamlBytes)
    yamlStr = UnquoteYAMLKey(yamlStr, "on")
    yamlStr = UnquoteYAMLKey(yamlStr, "if")
    yamlStr = CleanYAMLNullValues(yamlStr)
    
    return []byte(yamlStr), nil
}
```

### Example 2: Parsing Workflow Files

Parse YAML with error context:

```go
// pkg/parser/frontmatter_content.go
func ParseFrontmatter(content []byte) (map[string]any, error) {
    var data map[string]any
    
    // goccy/go-yaml provides detailed error context
    if err := yaml.Unmarshal(content, &data); err != nil {
        // Error includes line numbers and source snippets
        return nil, fmt.Errorf("invalid frontmatter: %w", err)
    }
    
    return data, nil
}
```

### Example 3: Nested Structure Ordering

Order nested job definitions:

```go
// Build a job with ordered fields
job := map[string]any{
    "steps": steps,
    "runs-on": "ubuntu-latest",
    "name": "Test Job",
    "permissions": perms,
}

// Order according to GitHub Actions conventions
orderedJob := OrderMapFields(job, []string{
    "name",
    "runs-on",
    "permissions",
    "environment",
    "if",
    "needs",
    "env",
    "steps",
})

// Add to jobs map
jobs := map[string]any{
    "test": orderedJob,
}
```

### Example 4: Custom Validation

Validate YAML structure before marshaling:

```go
func ValidateWorkflow(data map[string]any) error {
    // goccy/go-yaml validates during unmarshal
    // but you can add custom validation
    
    if _, ok := data["name"]; !ok {
        return fmt.Errorf("workflow name is required")
    }
    
    if _, ok := data["on"]; !ok {
        return fmt.Errorf("workflow triggers ('on') are required")
    }
    
    // Validate YAML can be marshaled
    _, err := yaml.Marshal(data)
    if err != nil {
        return fmt.Errorf("workflow contains invalid YAML: %w", err)
    }
    
    return nil
}
```

---

## Advanced Features

### YAML Path

Query and manipulate YAML using path expressions:

```go
import "github.com/goccy/go-yaml/parser"

content := `
name: Test Workflow
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout
`

file, _ := parser.ParseBytes([]byte(content), 0)
path, _ := yaml.PathString("$.jobs.test.runs-on")
node, _ := path.ReadNode(file)
fmt.Println(node.String()) // "ubuntu-latest"
```

### AST Manipulation

Access and modify the YAML Abstract Syntax Tree:

```go
import "github.com/goccy/go-yaml/parser"

file, _ := parser.ParseBytes([]byte(content), 0)

// Walk the AST
for _, doc := range file.Docs {
    for _, node := range doc.Body.(*ast.MappingNode).Values {
        key := node.Key.String()
        value := node.Value.String()
        fmt.Printf("%s: %s\n", key, value)
    }
}
```

### Custom Marshalers

Implement custom YAML marshaling:

```go
type WorkflowID string

func (w WorkflowID) MarshalYAML() ([]byte, error) {
    // Custom marshaling logic
    return yaml.Marshal(string(w))
}

func (w *WorkflowID) UnmarshalYAML(data []byte) error {
    // Custom unmarshaling logic
    var s string
    if err := yaml.Unmarshal(data, &s); err != nil {
        return err
    }
    *w = WorkflowID(s)
    return nil
}
```

### Strict Unmarshaling

Detect unknown fields during parsing:

```go
type WorkflowConfig struct {
    Name string `yaml:"name"`
    On   any    `yaml:"on"`
}

var config WorkflowConfig
decoder := yaml.NewDecoder(
    bytes.NewReader(content),
    yaml.Strict(), // Fail on unknown fields
)
err := decoder.Decode(&config)
```

---

## References

### Official Documentation

- **goccy/go-yaml GitHub**: https://github.com/goccy/go-yaml
- **Go Package Documentation**: https://pkg.go.dev/github.com/goccy/go-yaml
- **YAML Playground**: https://goccy.github.io/go-yaml/ (test YAML online)

### YAML Specifications

- **YAML 1.2 Spec**: https://yaml.org/spec/1.2/spec.html
- **YAML 1.1 Spec**: https://yaml.org/spec/1.1/
- **YAML Test Suite**: https://github.com/yaml/yaml-test-suite

### GitHub Actions

- **Workflow Syntax**: https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions
- **Best Practices**: https://docs.github.com/en/actions/learn-github-actions/usage-limits-billing-and-administration

### Related Documentation

- [yaml-version-gotchas.md](yaml-version-gotchas.md) - YAML 1.1 vs 1.2 compatibility issues
- [pkg/workflow/yaml.go](../pkg/workflow/yaml.go) - Field ordering utilities
- [pkg/workflow/yaml_options.go](../pkg/workflow/yaml_options.go) - Standard marshal options

### Library Features

The goccy/go-yaml library provides extensive features beyond basic marshaling:

- **Colored output** - Syntax highlighting for terminal display
- **Anchor/alias support** - YAML references and reuse
- **Comment preservation** - Maintain comments during round-trip
- **YAML Path** - JSONPath-like queries for YAML
- **Stream processing** - Handle large YAML files efficiently
- **Custom tags** - Define custom YAML types
- **Merge keys** - YAML `<<:` merge support

### Migration Notes

If migrating from other YAML libraries:

**From go-yaml/yaml (v2):**
- Change import: `gopkg.in/yaml.v2` → `github.com/goccy/go-yaml`
- API is similar but check error handling patterns
- Test with YAML 1.2 semantics (strings vs booleans)

**From go-yaml/yaml (v3):**
- Change import: `go.yaml.in/yaml/v3` → `github.com/goccy/go-yaml`
- API is very similar
- Benefit from better error messages
- YAML 1.1 → YAML 1.2 (check boolean keywords)

---

## Summary

GitHub Agentic Workflows uses `goccy/go-yaml` for its superior test coverage, excellent error messages, active maintenance, and YAML 1.2 compliance. The codebase demonstrates best practices including:

✅ **Deterministic ordering** via `yaml.MapSlice`  
✅ **Consistent options** via `DefaultMarshalOptions`  
✅ **Clean abstractions** (`MarshalWithFieldOrder`, `OrderMapFields`)  
✅ **GitHub Actions conventions** (field ordering, null handling)  
✅ **Proper error handling** with rich context

These patterns ensure that gh-aw generates professional, readable, and maintainable GitHub Actions workflows that follow community conventions.

---

**Last Updated**: 2026-02-16  
**Library Version**: goccy/go-yaml v1.19.1  
**Related Issues**: github/gh-aw#16054
