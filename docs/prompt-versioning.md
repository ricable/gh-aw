# Prompt Versioning Guide

This document describes the prompt versioning system in gh-aw and how to update prompt versions when making changes.

## Overview

gh-aw tracks versions for two types of prompts:

1. **System Prompts**: Built-in prompt files that provide instructions to AI agents (security policies, tool usage, etc.)
2. **Creator Prompts**: User-provided workflow markdown content

Both types are versioned and tracked in compiled workflow files for reproducibility, debugging, and auditing.

## System Prompt Versions

System prompts are located in:
- `actions/setup/md/*.md` - Runtime prompt files
- `pkg/workflow/prompts/*.md` - Compile-time embedded prompts

### Version Format

Versions use date-based format: `YYYY-MM-DD`

Example: `2026-02-17`

### Current System Prompts

| Prompt File | Purpose | Current Version |
|------------|---------|-----------------|
| `xpia.md` | Security policy and prohibited actions | `2026-02-17` |
| `temp_folder_prompt.md` | Temporary folder usage instructions | `2026-02-17` |
| `markdown.md` | Markdown generation guidelines | `2026-02-17` |
| `playwright_prompt.md` | Playwright tool output instructions | `2026-02-17` |
| `pr_context_prompt.md` | Pull request context instructions | `2026-02-17` |
| `cache_memory_prompt.md` | Cache memory usage (single cache) | `2026-02-17` |
| `cache_memory_prompt_multi.md` | Cache memory usage (multiple caches) | `2026-02-17` |
| `github_context_prompt.md` | GitHub context information | `2026-02-17` |
| `threat_detection.md` | Threat detection analysis prompt | `2026-02-17` |

## Updating Prompt Versions

When you modify a system prompt file, follow these steps:

### 1. Update the Prompt File

Add or update the version comment at the top of the file:

```markdown
<!-- Version: YYYY-MM-DD -->
```

Use today's date in `YYYY-MM-DD` format.

### 2. Update the Version Constant

Edit `pkg/workflow/prompt_versions.go` and update the corresponding constant:

```go
const (
    // XPIAPromptVersion is the version of the XPIA security policy prompt
    // File: actions/setup/md/xpia.md
    // Last modified: Brief description of changes
    XPIAPromptVersion PromptVersion = "YYYY-MM-DD"
    
    // ... other constants
)
```

### 3. Update the Version Manifest

The version manifest in `NewPromptVersionManifest()` should automatically pick up your constant change. Verify it maps correctly:

```go
func NewPromptVersionManifest() *PromptVersionManifest {
    return &PromptVersionManifest{
        GeneratedAt: time.Now(),
        SystemPrompts: map[string]PromptVersion{
            "xpia.md": XPIAPromptVersion,  // Should match your constant
            // ... other mappings
        },
    }
}
```

### 4. Rebuild and Recompile

```bash
make build       # Rebuild the binary with new versions
make recompile   # Recompile all workflows to include new versions
```

### 5. Verify Changes

Check a compiled workflow file to ensure the new version appears:

```bash
head -50 .github/workflows/typist.lock.yml | grep -A 10 "System Prompt Versions"
```

You should see your updated version in the output.

## Creator Prompt Hashing

Creator prompts (user workflow markdown) are automatically hashed during compilation. No manual updates needed.

The hash is:
- Computed from `WorkflowData.MarkdownContent`
- SHA-256 hash truncated to first 16 characters
- Included in the version manifest automatically

## Version Information in Workflows

Compiled workflows include version information in YAML comments:

```yaml
# frontmatter-hash: abc123...
#
# System Prompt Versions:
# Generated: 2026-02-17T12:00:00Z
# - cache_memory_prompt.md: 2026-02-17
# - cache_memory_prompt_multi.md: 2026-02-17
# - github_context_prompt.md: 2026-02-17
# - markdown.md: 2026-02-17
# - playwright_prompt.md: 2026-02-17
# - pr_context_prompt.md: 2026-02-17
# - temp_folder_prompt.md: 2026-02-17
# - threat_detection.md: 2026-02-17
# - xpia.md: 2026-02-17
# Creator Prompt Hash: 6091a0936f8d874e
```

## Best Practices

1. **Update versions when making substantial changes** - Minor typo fixes may not warrant a version bump
2. **Use today's date** - Versions reflect when changes were made, not when they were deployed
3. **Document changes** - Update the comment in the constant to describe what changed
4. **Test thoroughly** - Recompile and test workflows after version changes
5. **Commit together** - Commit prompt file and version constant changes in the same commit

## Troubleshooting

### Version not appearing in compiled workflows

1. Check that you updated the constant in `prompt_versions.go`
2. Verify the constant is correctly mapped in `NewPromptVersionManifest()`
3. Rebuild the binary: `make build`
4. Recompile workflows: `make recompile`

### Build errors after version update

1. Ensure the version string follows `YYYY-MM-DD` format
2. Check for syntax errors in `prompt_versions.go`
3. Run tests: `go test -v ./pkg/workflow/`

### Inconsistent versions across workflows

1. Make sure all workflows were recompiled after the version update
2. Run `make recompile` to regenerate all `.lock.yml` files
3. Check for any compilation errors that may have skipped some workflows

## Related Files

- `pkg/workflow/prompt_versions.go` - Version constants and manifest
- `pkg/workflow/prompt_versions_test.go` - Unit tests for versioning
- `pkg/workflow/compiler_yaml.go` - Integration with workflow compilation
- `actions/setup/md/*.md` - System prompt files
- `pkg/workflow/prompts/*.md` - Embedded prompt files

## Future Enhancements

Potential improvements to the versioning system:

1. **Automatic version bumping** - Detect prompt file changes and auto-update versions
2. **Version diff reporting** - Show which prompts changed between compilations
3. **Historical tracking** - Maintain a changelog of prompt versions
4. **Version validation** - Ensure version format consistency at compile time
