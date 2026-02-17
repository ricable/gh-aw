#!/usr/bin/env bash
set -euo pipefail

VERSION_FILE=".github/aw/prompt-pack-version.txt"

if [[ ! -f "$VERSION_FILE" ]]; then
  echo "❌ Missing version file: $VERSION_FILE"
  echo "Create it with a semantic version (for example: 0.1.0)."
  exit 1
fi

version_value="$(tr -d '[:space:]' < "$VERSION_FILE")"
if [[ ! "$version_value" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
  echo "❌ Invalid prompt pack version in $VERSION_FILE: '$version_value'"
  echo "Expected SemVer format: MAJOR.MINOR.PATCH (for example: 1.2.3)."
  exit 1
fi

if [[ -n "${GITHUB_BASE_REF:-}" ]]; then
  git fetch --no-tags origin "$GITHUB_BASE_REF" >/dev/null 2>&1 || true
  base_ref="origin/$GITHUB_BASE_REF"
  if git rev-parse --verify "$base_ref" >/dev/null 2>&1; then
    base_commit="$(git merge-base HEAD "$base_ref")"
  else
    base_commit="HEAD~1"
  fi
else
  base_commit="HEAD~1"
fi

if ! git rev-parse --verify "$base_commit" >/dev/null 2>&1; then
  echo "ℹ️ Unable to determine base commit, skipping prompt version check"
  exit 0
fi

changed_files="$(git diff --name-only "$base_commit"...HEAD)"

if [[ -z "$changed_files" ]]; then
  echo "ℹ️ No changed files detected, skipping prompt version check"
  exit 0
fi

# Prompt source files that materially affect prompt behavior/routing.
# Keep this list explicit to avoid forcing version bumps for docs-only .github/aw edits.
prompt_source_patterns=(
  ".github/agents/agentic-workflows.agent.md"
  ".github/aw/create-agentic-workflow.md"
  ".github/aw/update-agentic-workflow.md"
  ".github/aw/debug-agentic-workflow.md"
  ".github/aw/upgrade-agentic-workflows.md"
  ".github/aw/create-shared-agentic-workflow.md"
  ".github/aw/orchestration.md"
  ".github/aw/projects.md"
  ".github/aw/github-agentic-workflows.md"
  ".github/aw/serena-tool.md"
  "actions/setup/md/*.md"
  "pkg/workflow/prompts/*.md"
)

prompt_changed_files=()
while IFS= read -r file; do
  [[ -z "$file" ]] && continue
  for pattern in "${prompt_source_patterns[@]}"; do
    if [[ "$file" == $pattern ]]; then
      prompt_changed_files+=("$file")
      break
    fi
  done
done <<< "$changed_files"

if [[ ${#prompt_changed_files[@]} -eq 0 ]]; then
  echo "✅ No prompt source changes detected"
  exit 0
fi

version_file_changed=false
while IFS= read -r file; do
  if [[ "$file" == "$VERSION_FILE" ]]; then
    version_file_changed=true
    break
  fi
done <<< "$changed_files"

if [[ "$version_file_changed" == true ]]; then
  echo "✅ Prompt source changes detected and version file was updated ($VERSION_FILE: $version_value)"
  exit 0
fi

echo "❌ Prompt source changes detected without a prompt pack version bump"
echo ""
echo "Changed prompt source files:"
printf '  - %s\n' "${prompt_changed_files[@]}"
echo ""
echo "Update $VERSION_FILE (current: $version_value) with a SemVer bump:"
echo "  - MAJOR: behavior/routing/security changes"
echo "  - MINOR: new capabilities/sections"
echo "  - PATCH: wording/clarity changes"
exit 1
