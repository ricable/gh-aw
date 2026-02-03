# MCP Server Go Tool Refactoring

## Overview

This document describes the refactoring of the MCP server Go command "add" tool from calling the `gh aw add` executable to calling the Go add command function directly.

## Problem

The original implementation would execute `gh aw add` as a subprocess using `os/exec`, which:
- Added subprocess overhead
- Made error handling more complex
- Couldn't access structured return values
- Required shell command parsing

## Solution

Refactored the tool to call `cli.AddWorkflows()` directly from the Go code.

## Implementation Details

### File Location
`.github/workflows/shared/add-workflow-safe-input.md`

### Before

```go
import (
  "context"
  "os/exec"
)

// Build gh aw add command
args := []string{"aw", "add", workflow}
if engine != "" {
  args = append(args, "--engine", engine)
}

// Execute gh aw add command
cmd := exec.CommandContext(context.Background(), "gh", args...)
cmd.Run()
```

### After

```go
import (
  "github.com/github/gh-aw/pkg/cli"
)

// Call AddWorkflows function directly
addResult, err := cli.AddWorkflows(
  workflows,
  1,              // number of copies
  verbose,        // verbose flag
  engine,         // engine override
  "",             // name (empty for default)
  force,          // force flag
  "",             // append text
  false,          // create PR
  false,          // push
  false,          // no gitattributes
  "",             // workflow dir
  false,          // no stop-after
  "",             // stop-after value
)
```

## Benefits

1. **Performance**: Eliminates subprocess spawning overhead
2. **Simplicity**: Direct function call is cleaner and more maintainable
3. **Error Handling**: Better error propagation without parsing stderr
4. **Return Values**: Access to structured result including:
   - PR number
   - PR URL
   - Workflow dispatch status

## Usage

The tool is accessible as `safeinputs-add` (or `safeinputs_add` after normalization):

```
safeinputs-add with workflow: "owner/repo/workflow-name"
safeinputs-add with workflow: "owner/repo/workflow-name@v1.0.0", engine: "copilot"
safeinputs-add with workflow: "owner/repo/workflow-name", force: true, verbose: true
```

## Testing

Tested with:
- Workflow compilation: `./gh-aw compile test-add-safe-input`
- Safe-inputs tests: All passing
- Add command tests: All passing
- Full recompilation: All 150 workflows compiled successfully

## Related Files

- `.github/workflows/shared/add-workflow-safe-input.md` - Safe-input tool definition
- `.github/workflows/test-add-safe-input.md` - Test workflow
- `pkg/cli/add_command.go` - Add command implementation
