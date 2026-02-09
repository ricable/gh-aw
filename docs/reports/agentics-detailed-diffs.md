# Detailed Diffs for Agentics Collection Fixes

This document shows the exact changes made to each workflow file.

## Summary of Changes

- **9 workflows** with deprecated `bash:` syntax
- **5 workflows** with deprecated `add-comment.discussion` field

---

## 1. daily-backlog-burner.md

### Changes:
- Removed `discussion: true` from `add-comment` config
- Changed `bash:` to `bash: true`

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

---

## 2. daily-dependency-updates.md

### Changes:
- Changed `bash:` to `bash: true`

```diff
 tools:
   web-fetch:
   github:
     toolsets: [all]
-  bash:
+  bash: true
```

---

## 3. daily-perf-improver.md

### Changes:
- Removed `discussion: true` from `add-comment` config
- Changed `bash:` to `bash: true`

```diff
 safe-outputs:
   add-comment:
-    discussion: true
     target: "*"

 tools:
-  bash:
+  bash: true
```

---

## 4. daily-progress.md

### Changes:
- Changed `bash:` to `bash: true`

```diff
 tools:
   web-fetch:
   github:
     toolsets: [all]
-  bash:
+  bash: true
```

---

## 5. daily-qa.md

### Changes:
- Removed `discussion: true` from `add-comment` config
- Changed `bash:` to `bash: true`

```diff
 safe-outputs:
   add-comment:
-    discussion: true
     target: "*"

 tools:
-  bash:
+  bash: true
```

---

## 6. daily-test-improver.md

### Changes:
- Removed `discussion: true` from `add-comment` config
- Changed `bash:` to `bash: true`

```diff
 safe-outputs:
   add-comment:
-    discussion: true
     target: "*"

 tools:
-  bash:
+  bash: true
```

---

## 7. pr-fix.md

### Changes:
- Changed `bash:` to `bash: true`

```diff
 tools:
   web-fetch:
   github:
     toolsets: [all]
-  bash:
+  bash: true
```

---

## 8. q.md

### Changes:
- Changed `bash:` to `bash: true`

```diff
 tools:
-  bash:
+  bash: true
```

---

## 9. repo-ask.md

### Changes:
- Changed `bash:` to `bash: true`

```diff
 tools:
   web-fetch:
   github:
     toolsets: [all]
-  bash:
+  bash: true
```

---

## 10. daily-accessibility-review.md

### Changes:
- Removed `discussion: true` from `add-comment` config

```diff
 safe-outputs:
   add-comment:
-    discussion: true
```

---

## Workflows Without Changes (9 files)

The following workflows already use current syntax and required no changes:

1. ci-doctor.md
2. daily-plan.md
3. daily-repo-status.md
4. daily-team-status.md
5. issue-triage.md
6. plan.md
7. update-docs.md
8. weekly-research.md
9. shared/reporting.md

---

## How These Fixes Were Applied

All fixes were applied automatically using the `gh aw fix` command:

```bash
cd agentics
gh aw fix --write
```

The following codemods were automatically applied:
- `bash-anonymous-removal`: Replaces `bash:` with `bash: true`
- `discussion-flag-removal`: Removes deprecated `add-comment.discussion` field

## Verification

After applying fixes, all workflows compile successfully:

```bash
gh aw compile --validate --strict
✓ Compiled 18 workflow(s): 0 error(s), 11 warning(s)
```
