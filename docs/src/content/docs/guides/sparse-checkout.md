# User Guide: Sparse Checkout with Runtime Imports

## For Users Encountering the Issue

If you're seeing this error:
```
Error: Failed to process runtime import for .github/workflows/your-workflow.md: Runtime import file not found: workflows/your-workflow.md
```

And your workflow uses `git sparse-checkout` in custom steps, you have two options:

### Option 1: Upgrade gh-aw (Recommended)

Upgrade to gh-aw version that includes the sparse-checkout fix (version >= the one containing this fix). The compiler will automatically add a recovery step to re-checkout `.github` and `.agents` folders.

```bash
# Update gh-aw
gh extension upgrade gh-aw

# Recompile your workflow
gh aw compile .github/workflows/your-workflow.md
```

The compiled workflow will now include an automatic recovery step before runtime imports are processed.

### Option 2: Manual Workaround (If Upgrade Not Possible)

Add a manual step to your workflow frontmatter to re-checkout `.github` after sparse-checkout:

```yaml
---
steps:
  - name: Checkout specific files
    run: |
      git sparse-checkout init --cone
      git sparse-checkout set src
  
  # Add this step manually to restore .github folder
  - name: Re-checkout workflow files
    uses: actions/checkout@v6
    with:
      sparse-checkout: |
        .github
        .agents
      fetch-depth: 1
      persist-credentials: false
---
```

## Best Practices

### ✅ Recommended: Include .github in sparse-checkout

If you need to use sparse-checkout, include `.github` in your checkout pattern:

```yaml
steps:
  - name: Checkout files
    run: |
      git sparse-checkout init --cone
      git sparse-checkout set src .github .agents
```

This avoids the issue entirely and ensures workflow files are always available.

### ⚠️ Alternative: Let the compiler handle it

With the fix applied, you can continue using sparse-checkout without `.github`, and the compiler will automatically add a recovery step. However, this adds an extra checkout step to your workflow.

## How It Works

The gh-aw compiler now:

1. **Detects** when custom steps use `git sparse-checkout` commands
2. **Checks** if the workflow uses runtime imports (which need `.github` folder)
3. **Inserts** a recovery step automatically before prompt generation:
   ```yaml
   - name: Re-checkout .github and .agents after sparse-checkout
     uses: actions/checkout@v6
     with:
       sparse-checkout: |
         .github
         .agents
       fetch-depth: 1
       persist-credentials: false
   ```

This step runs after your custom sparse-checkout steps but before runtime imports are processed, ensuring workflow files are available.

## Performance Considerations

The recovery step adds a minimal overhead:
- Shallow clone (fetch-depth: 1)
- Only checks out `.github` and `.agents` directories
- Typical execution time: 1-3 seconds

For optimal performance, prefer including `.github` in your initial sparse-checkout pattern if possible.

## Verification

To verify the fix is working:

1. Compile your workflow:
   ```bash
   gh aw compile .github/workflows/your-workflow.md
   ```

2. Check the lock file for the recovery step:
   ```bash
   grep "Re-checkout .github and .agents" .github/workflows/your-workflow.lock.yml
   ```

3. Verify step order:
   ```bash
   grep "^ *- name:" .github/workflows/your-workflow.lock.yml
   ```

You should see the recovery step between your custom sparse-checkout steps and the "Create prompt" step.

## Related Issues

- Original issue: a3-python regression in Z3Prover/z3 repository
- Fix PR: github/gh-aw#[PR_NUMBER]
- Documentation: scratchpad/sparse-checkout-fix.md
