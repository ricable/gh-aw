// @ts-check
/// <reference types="@actions/github-script" />

const fs = require("fs");
const path = require("path");

const { getBaseBranch } = require("./get_base_branch.cjs");
const { getErrorMessage } = require("./error_helpers.cjs");
const { execGitSync } = require("./git_helpers.cjs");

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

  if (typeof core !== "undefined" && core.info) {
    core.info(`[generate_git_patch] branchName="${branchName || ""}" GITHUB_SHA="${githubSha || ""}" defaultBranch="${defaultBranch}"`);
  }

  // Ensure /tmp/gh-aw directory exists
  const patchDir = path.dirname(patchPath);
  if (!fs.existsSync(patchDir)) {
    fs.mkdirSync(patchDir, { recursive: true });
  }

  let patchGenerated = false;
  let errorMessage = null;

  try {
    // Strategy 1: If we have a branch name, check if that branch exists and get its diff
    if (branchName) {
      if (typeof core !== "undefined" && core.info) {
        core.info(`[generate_git_patch] Strategy 1: branch-based patch for "${branchName}"`);
      }
      // Check if the branch exists locally
      try {
        execGitSync(["show-ref", "--verify", "--quiet", `refs/heads/${branchName}`], { cwd });

        // Determine base ref for patch generation
        let baseRef;
        try {
          // Check if origin/branchName exists in remote tracking refs
          execGitSync(["show-ref", "--verify", "--quiet", `refs/remotes/origin/${branchName}`], { cwd });
          baseRef = `origin/${branchName}`;
          if (typeof core !== "undefined" && core.info) {
            core.info(`[generate_git_patch] using remote tracking ref as baseRef="${baseRef}"`);
          }
        } catch {
          // Remote tracking ref not found (e.g. after gh pr checkout which doesn't set tracking refs).
          // Try fetching the branch from origin so we use only the NEW commits as the patch base.
          if (typeof core !== "undefined" && core.info) {
            core.info(`[generate_git_patch] refs/remotes/origin/${branchName} not found; fetching from origin`);
          }
          try {
            execGitSync(["fetch", "origin", branchName], { cwd });
            baseRef = `origin/${branchName}`;
            if (typeof core !== "undefined" && core.info) {
              core.info(`[generate_git_patch] fetch succeeded, baseRef="${baseRef}"`);
            }
          } catch (fetchErr) {
            // Branch doesn't exist on origin yet (new branch) – fall back to merge-base
            if (typeof core !== "undefined" && core.warning) {
              core.warning(`[generate_git_patch] fetch of origin/${branchName} failed (${getErrorMessage(fetchErr)}); falling back to merge-base with "${defaultBranch}"`);
            }
            execGitSync(["fetch", "origin", defaultBranch], { cwd });
            baseRef = execGitSync(["merge-base", `origin/${defaultBranch}`, branchName], { cwd }).trim();
            if (typeof core !== "undefined" && core.info) {
              core.info(`[generate_git_patch] merge-base baseRef="${baseRef}"`);
            }
          }
        }

        // Count commits to be included
        const commitCount = parseInt(execGitSync(["rev-list", "--count", `${baseRef}..${branchName}`], { cwd }).trim(), 10);
        if (typeof core !== "undefined" && core.info) {
          core.info(`[generate_git_patch] ${commitCount} commit(s) between ${baseRef} and ${branchName}`);
        }

        if (commitCount > 0) {
          // Generate patch from the determined base to the branch
          const patchContent = execGitSync(["format-patch", `${baseRef}..${branchName}`, "--stdout"], { cwd });

          if (patchContent && patchContent.trim()) {
            fs.writeFileSync(patchPath, patchContent, "utf8");
            if (typeof core !== "undefined" && core.info) {
              core.info(`[generate_git_patch] patch written: ${patchContent.split("\n").length} lines, ${Math.ceil(Buffer.byteLength(patchContent, "utf8") / 1024)} KB`);
            }
            patchGenerated = true;
          } else {
            if (typeof core !== "undefined" && core.warning) {
              core.warning(`[generate_git_patch] format-patch produced empty output for ${baseRef}..${branchName}`);
            }
          }
        } else {
          if (typeof core !== "undefined" && core.info) {
            core.info(`[generate_git_patch] no commits to patch (Strategy 1)`);
          }
        }
      } catch (branchError) {
        // Branch does not exist locally
        if (typeof core !== "undefined" && core.info) {
          core.info(`[generate_git_patch] local branch "${branchName}" not found: ${getErrorMessage(branchError)}`);
        }
      }
    } else {
      if (typeof core !== "undefined" && core.info) {
        core.info(`[generate_git_patch] Strategy 1: skipped (no branchName)`);
      }
    }

    // Strategy 2: Check if commits were made to current HEAD since checkout
    if (!patchGenerated) {
      const currentHead = execGitSync(["rev-parse", "HEAD"], { cwd }).trim();
      if (typeof core !== "undefined" && core.info) {
        core.info(`[generate_git_patch] Strategy 2: HEAD="${currentHead}" GITHUB_SHA="${githubSha || ""}"`);
      }

      if (!githubSha) {
        errorMessage = "GITHUB_SHA environment variable is not set";
        if (typeof core !== "undefined" && core.warning) {
          core.warning(`[generate_git_patch] ${errorMessage}`);
        }
      } else if (currentHead === githubSha) {
        // No commits have been made since checkout
        if (typeof core !== "undefined" && core.info) {
          core.info(`[generate_git_patch] HEAD matches GITHUB_SHA – no new commits`);
        }
      } else {
        // Check if GITHUB_SHA is an ancestor of current HEAD
        try {
          execGitSync(["merge-base", "--is-ancestor", githubSha, "HEAD"], { cwd });

          // Count commits between GITHUB_SHA and HEAD
          const commitCount = parseInt(execGitSync(["rev-list", "--count", `${githubSha}..HEAD`], { cwd }).trim(), 10);
          if (typeof core !== "undefined" && core.info) {
            core.info(`[generate_git_patch] ${commitCount} commit(s) between GITHUB_SHA and HEAD`);
          }

          if (commitCount > 0) {
            // Generate patch from GITHUB_SHA to HEAD
            const patchContent = execGitSync(["format-patch", `${githubSha}..HEAD`, "--stdout"], { cwd });

            if (patchContent && patchContent.trim()) {
              fs.writeFileSync(patchPath, patchContent, "utf8");
              if (typeof core !== "undefined" && core.info) {
                core.info(`[generate_git_patch] patch written: ${patchContent.split("\n").length} lines, ${Math.ceil(Buffer.byteLength(patchContent, "utf8") / 1024)} KB`);
              }
              patchGenerated = true;
            } else {
              if (typeof core !== "undefined" && core.warning) {
                core.warning(`[generate_git_patch] format-patch produced empty output for ${githubSha}..HEAD`);
              }
            }
          } else {
            if (typeof core !== "undefined" && core.info) {
              core.info(`[generate_git_patch] no commits to patch (Strategy 2)`);
            }
          }
        } catch {
          // GITHUB_SHA is not an ancestor of HEAD - repository state has diverged
          if (typeof core !== "undefined" && core.warning) {
            core.warning(`[generate_git_patch] GITHUB_SHA is not an ancestor of HEAD – repository state has diverged`);
          }
        }
      }
    }
  } catch (error) {
    errorMessage = `Failed to generate patch: ${getErrorMessage(error)}`;
    if (typeof core !== "undefined" && core.warning) {
      core.warning(`[generate_git_patch] ${errorMessage}`);
    }
  }

  // Check if patch was generated and has content
  if (patchGenerated && fs.existsSync(patchPath)) {
    const patchContent = fs.readFileSync(patchPath, "utf8");
    const patchSize = Buffer.byteLength(patchContent, "utf8");
    const patchLines = patchContent.split("\n").length;

    if (!patchContent.trim()) {
      // Empty patch
      return {
        success: false,
        error: "No changes to commit - patch is empty",
        patchPath: patchPath,
        patchSize: 0,
        patchLines: 0,
      };
    }

    return {
      success: true,
      patchPath: patchPath,
      patchSize: patchSize,
      patchLines: patchLines,
    };
  }

  // No patch generated
  return {
    success: false,
    error: errorMessage || "No changes to commit - no commits found",
    patchPath: patchPath,
  };
}

module.exports = {
  generateGitPatch,
};
