// @ts-check
/// <reference types="@actions/github-script" />

const fs = require("fs");
const path = require("path");

const { getBaseBranch } = require("./get_base_branch.cjs");
const { getErrorMessage } = require("./error_helpers.cjs");
const { execGitSync } = require("./git_helpers.cjs");

/**
 * Log an informational message to stderr with a timestamp.
 * When the GitHub Actions `core` global is available the message is also
 * forwarded to `core.info` so it appears in the Actions log.
 * @param {string} msg
 */
function logInfo(msg) {
  const timestamp = new Date().toISOString();
  process.stderr.write(`[${timestamp}] [generate_git_patch] ${msg}\n`);
  if (typeof core !== "undefined" && core.info) {
    core.info(`[generate_git_patch] ${msg}`);
  }
}

/**
 * Log a warning message to stderr with a timestamp.
 * When the GitHub Actions `core` global is available the message is also
 * forwarded to `core.warning`.
 * @param {string} msg
 */
function logWarning(msg) {
  const timestamp = new Date().toISOString();
  process.stderr.write(`[${timestamp}] [generate_git_patch] WARNING: ${msg}\n`);
  if (typeof core !== "undefined" && core.warning) {
    core.warning(`[generate_git_patch] ${msg}`);
  }
}

/**
 * Generates a git patch file for the current changes
 * @param {string} branchName - The branch name to generate patch for
 * @returns {Object} Object with patch info or error
 */
function generateGitPatch(branchName) {
  const patchPath = "/tmp/gh-aw/aw.patch";
  const cwd = process.env.GITHUB_WORKSPACE || process.cwd();
  const defaultBranch = process.env.DEFAULT_BRANCH || getBaseBranch();
  const githubSha = process.env.GITHUB_SHA;

  logInfo(`Starting patch generation`);
  logInfo(`  branchName  : ${branchName || "(none)"}`);
  logInfo(`  GITHUB_SHA  : ${githubSha || "(not set)"}`);
  logInfo(`  defaultBranch: ${defaultBranch}`);
  logInfo(`  cwd         : ${cwd}`);
  logInfo(`  patchPath   : ${patchPath}`);

  // Ensure /tmp/gh-aw directory exists
  const patchDir = path.dirname(patchPath);
  if (!fs.existsSync(patchDir)) {
    fs.mkdirSync(patchDir, { recursive: true });
    logInfo(`Created patch directory: ${patchDir}`);
  }

  let patchGenerated = false;
  let errorMessage = null;

  try {
    // Strategy 1: If we have a branch name, check if that branch exists and get its diff
    if (branchName) {
      logInfo(`Strategy 1: attempting branch-based patch for "${branchName}"`);
      // Check if the branch exists locally
      try {
        execGitSync(["show-ref", "--verify", "--quiet", `refs/heads/${branchName}`], { cwd });
        logInfo(`  Local branch "${branchName}" exists`);

        // Determine base ref for patch generation
        let baseRef;
        try {
          // Check if origin/branchName exists in remote tracking refs
          execGitSync(["show-ref", "--verify", "--quiet", `refs/remotes/origin/${branchName}`], { cwd });
          baseRef = `origin/${branchName}`;
          logInfo(`  Remote tracking ref refs/remotes/origin/${branchName} found → baseRef="${baseRef}"`);
        } catch {
          // Remote tracking ref not found (e.g. after gh pr checkout which doesn't set tracking refs).
          // Try fetching the branch from origin so we use only the NEW commits as the patch base.
          logInfo(`  refs/remotes/origin/${branchName} not found; attempting "git fetch origin ${branchName}"`);
          try {
            execGitSync(["fetch", "origin", branchName], { cwd });
            baseRef = `origin/${branchName}`;
            logInfo(`  Fetch succeeded → baseRef="${baseRef}"`);
          } catch (fetchErr) {
            // Branch doesn't exist on origin yet (new branch) – fall back to merge-base
            logWarning(`  Fetch of origin/${branchName} failed (${getErrorMessage(fetchErr)}); falling back to merge-base with "${defaultBranch}"`);
            execGitSync(["fetch", "origin", defaultBranch], { cwd });
            baseRef = execGitSync(["merge-base", `origin/${defaultBranch}`, branchName], { cwd }).trim();
            logInfo(`  merge-base result → baseRef="${baseRef}"`);
          }
        }

        // Count commits to be included
        const commitCount = parseInt(execGitSync(["rev-list", "--count", `${baseRef}..${branchName}`], { cwd }).trim(), 10);
        logInfo(`  Commits between ${baseRef} and ${branchName}: ${commitCount}`);

        if (commitCount > 0) {
          // List each commit SHA for traceability
          try {
            const commitList = execGitSync(["log", "--oneline", `${baseRef}..${branchName}`], { cwd }).trim();
            logInfo(
              `  Commits to include:\n${commitList
                .split("\n")
                .map(l => `    ${l}`)
                .join("\n")}`
            );
          } catch {
            // Non-fatal – best-effort logging
          }

          // Generate patch from the determined base to the branch
          logInfo(`  Generating patch: git format-patch ${baseRef}..${branchName} --stdout`);
          const patchContent = execGitSync(["format-patch", `${baseRef}..${branchName}`, "--stdout"], { cwd });

          if (patchContent && patchContent.trim()) {
            fs.writeFileSync(patchPath, patchContent, "utf8");
            const patchSizeKb = Math.ceil(Buffer.byteLength(patchContent, "utf8") / 1024);
            const patchLines = patchContent.split("\n").length;
            logInfo(`  Patch written: ${patchLines} lines, ${patchSizeKb} KB`);
            patchGenerated = true;
          } else {
            logWarning(`  format-patch produced empty output for ${baseRef}..${branchName}`);
          }
        } else {
          logInfo(`  No commits to patch (commitCount=0); skipping Strategy 1`);
        }
      } catch (branchError) {
        // Branch does not exist locally
        logInfo(`  Local branch "${branchName}" not found (${getErrorMessage(branchError)}); skipping Strategy 1`);
      }
    } else {
      logInfo(`Strategy 1: skipped (no branchName provided)`);
    }

    // Strategy 2: Check if commits were made to current HEAD since checkout
    if (!patchGenerated) {
      logInfo(`Strategy 2: checking for commits on current HEAD since GITHUB_SHA`);
      const currentHead = execGitSync(["rev-parse", "HEAD"], { cwd }).trim();
      logInfo(`  currentHead : ${currentHead}`);
      logInfo(`  GITHUB_SHA  : ${githubSha || "(not set)"}`);

      if (!githubSha) {
        errorMessage = "GITHUB_SHA environment variable is not set";
        logWarning(`  ${errorMessage}`);
      } else if (currentHead === githubSha) {
        // No commits have been made since checkout
        logInfo(`  HEAD matches GITHUB_SHA – no new commits since checkout`);
      } else {
        logInfo(`  HEAD differs from GITHUB_SHA – checking ancestry`);
        // Check if GITHUB_SHA is an ancestor of current HEAD
        try {
          execGitSync(["merge-base", "--is-ancestor", githubSha, "HEAD"], { cwd });
          logInfo(`  GITHUB_SHA is an ancestor of HEAD`);

          // Count commits between GITHUB_SHA and HEAD
          const commitCount = parseInt(execGitSync(["rev-list", "--count", `${githubSha}..HEAD`], { cwd }).trim(), 10);
          logInfo(`  Commits between GITHUB_SHA and HEAD: ${commitCount}`);

          if (commitCount > 0) {
            // List each commit SHA for traceability
            try {
              const commitList = execGitSync(["log", "--oneline", `${githubSha}..HEAD`], { cwd }).trim();
              logInfo(
                `  Commits to include:\n${commitList
                  .split("\n")
                  .map(l => `    ${l}`)
                  .join("\n")}`
              );
            } catch {
              // Non-fatal
            }

            // Generate patch from GITHUB_SHA to HEAD
            logInfo(`  Generating patch: git format-patch ${githubSha}..HEAD --stdout`);
            const patchContent = execGitSync(["format-patch", `${githubSha}..HEAD`, "--stdout"], { cwd });

            if (patchContent && patchContent.trim()) {
              fs.writeFileSync(patchPath, patchContent, "utf8");
              const patchSizeKb = Math.ceil(Buffer.byteLength(patchContent, "utf8") / 1024);
              const patchLines = patchContent.split("\n").length;
              logInfo(`  Patch written: ${patchLines} lines, ${patchSizeKb} KB`);
              patchGenerated = true;
            } else {
              logWarning(`  format-patch produced empty output for ${githubSha}..HEAD`);
            }
          } else {
            logInfo(`  No commits to patch (commitCount=0)`);
          }
        } catch {
          // GITHUB_SHA is not an ancestor of HEAD - repository state has diverged
          logWarning(`  GITHUB_SHA is NOT an ancestor of HEAD – repository state has diverged; cannot generate patch`);
        }
      }
    }
  } catch (error) {
    errorMessage = `Failed to generate patch: ${getErrorMessage(error)}`;
    logWarning(errorMessage);
  }

  // Check if patch was generated and has content
  if (patchGenerated && fs.existsSync(patchPath)) {
    const patchContent = fs.readFileSync(patchPath, "utf8");
    const patchSize = Buffer.byteLength(patchContent, "utf8");
    const patchLines = patchContent.split("\n").length;

    if (!patchContent.trim()) {
      // Empty patch
      logWarning(`Patch file exists but is empty`);
      return {
        success: false,
        error: "No changes to commit - patch is empty",
        patchPath: patchPath,
        patchSize: 0,
        patchLines: 0,
      };
    }

    logInfo(`Patch generation succeeded: ${patchLines} lines, ${Math.ceil(patchSize / 1024)} KB`);
    return {
      success: true,
      patchPath: patchPath,
      patchSize: patchSize,
      patchLines: patchLines,
    };
  }

  // No patch generated
  const finalError = errorMessage || "No changes to commit - no commits found";
  logInfo(`Patch generation result: no patch – ${finalError}`);
  return {
    success: false,
    error: finalError,
    patchPath: patchPath,
  };
}

module.exports = {
  generateGitPatch,
};
