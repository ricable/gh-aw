# Agentics Collection Syntax Check Report

**Date:** 2026-02-09  
**Repository Analyzed:** https://github.com/githubnext/agentics  
**Branch:** main  
**Total Workflows:** 18

## Executive Summary

All 18 workflows in the agentics collection have been checked for latest gh-aw syntax. **9 workflows** had deprecated syntax that was automatically fixed using `gh aw fix` command. All workflows now compile successfully with strict validation enabled.

## Issues Found

### 1. Deprecated `bash:` Anonymous Syntax (9 workflows)

**Issue:** Workflows were using `bash:` without an explicit value, which is deprecated syntax.

**Affected Workflows:**
1. daily-backlog-burner.md
2. daily-dependency-updates.md  
3. daily-perf-improver.md
4. daily-progress.md
5. daily-qa.md
6. daily-test-improver.md
7. pr-fix.md
8. q.md
9. repo-ask.md

**Error Message:**
```
invalid bash tool configuration: anonymous syntax 'bash:' is not supported. 
Use 'bash: true' (enable all commands), 'bash: false' (disable), or 
'bash: ["cmd1", "cmd2"]' (specific commands). Run 'gh aw fix' to automatically migrate
```

**Fix Applied:**
```diff
tools:
-  bash:
+  bash: true
```

### 2. Deprecated `add-comment.discussion` Field (4 workflows)

**Issue:** Workflows were using the deprecated `add-comment.discussion` field.

**Affected Workflows:**
1. daily-accessibility-review.md
2. daily-backlog-burner.md
3. daily-perf-improver.md
4. daily-qa.md
5. daily-test-improver.md

**Fix Applied:**
```diff
safe-outputs:
  add-comment:
-    discussion: true
    target: "*"
    max: 3
```

## Workflows Without Issues (9 workflows)

The following workflows already use current syntax:
1. ci-doctor.md
2. daily-plan.md
3. daily-repo-status.md
4. daily-team-status.md
5. issue-triage.md
6. plan.md
7. update-docs.md
8. weekly-research.md
9. shared/reporting.md (shared import file)

## Verification Results

After applying fixes with `gh aw fix --write`:

```
✓ ✓ Fixed 9 of 18 workflow files
✓ Compiled 18 workflow(s): 0 error(s), 11 warning(s)
```

All workflows now compile successfully. The warnings are expected:
- "Unable to pin action" warnings (no GitHub API access in test environment)
- "Fuzzy schedule scattering" warnings (no git remote configured in test environment)

## Automatic Fix Tool

The deprecated syntax can be automatically fixed using:

```bash
gh aw fix --write
```

This applies the following codemods:
- `bash-anonymous-removal`: Replaces `bash:` with `bash: true`
- `discussion-flag-removal`: Removes deprecated `add-comment.discussion` field

## Recommendations

1. **Apply fixes to agentics repository:** Run `gh aw fix --write` in the agentics repository to update all workflows
2. **Recompile workflows:** Run `gh aw compile` after fixing to regenerate `.lock.yml` files
3. **Regular syntax checks:** Add periodic checks (monthly) to ensure workflows stay up-to-date with latest syntax
4. **CI validation:** Consider adding `gh aw compile --validate --strict` to CI pipeline

## Sample Diff

Example of changes applied to `daily-backlog-burner.md`:

```diff
 safe-outputs:
   add-comment:
-    discussion: true
     target: "*"
     max: 3

 tools:
   web-fetch:
   github:
     toolsets: [all]
-  bash:
+  bash: true
```

## Next Steps

1. Create pull request in agentics repository with these fixes
2. Request review from agentics maintainers
3. Merge and deploy updated workflows
4. Monitor workflow runs to ensure no regressions

## References

- gh-aw documentation: https://github.com/github/gh-aw
- Fix command help: `gh aw fix --help`
- Compile command help: `gh aw compile --help`
