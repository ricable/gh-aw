// @ts-check
/// <reference types="@actions/github-script" />

const fs = require("fs");
const path = require("path");

const { getBaseBranch } = require("./get_base_branch.cjs");
const { getErrorMessage } = require("./error_helpers.cjs");
const { execGitSync } = require("./git_helpers.cjs");

const PATCH_PATH = "/tmp/gh-aw/aw.patch";

/**
 * Resolves the base ref to use for patch generation against a named branch.
 * Preference order:
 *   1. Remote tracking ref  refs/remotes/origin/<branch>  (already fetched)
 *   2. Fresh fetch of origin/<branch>                     (gh pr checkout path)
 *   3. merge-base with origin/<defaultBranch>             (brand-new branch)
 * @param {string} branchName
 * @param {string} defaultBranch
 * @param {string} cwd
 * @returns {string} baseRef
 */
function resolveBaseRef(branchName, defaultBranch, cwd) {
  try {
    execGitSync(["show-ref", "--verify", "--quiet", `refs/remotes/origin/${branchName}`], { cwd });
    const baseRef = `origin/${branchName}`;
    core.info(`[generate_git_patch] using remote tracking ref as baseRef="${baseRef}"`);
    return baseRef;
  } catch {
    // Remote tracking ref not found (e.g. after gh pr checkout which doesn't set tracking refs).
    // Try fetching the branch from origin so we use only the NEW commits as the patch base.
    core.info(`[generate_git_patch] refs/remotes/origin/${branchName} not found; fetching from origin`);
  }

  try {
    execGitSync(["fetch", "origin", branchName], { cwd });
    const baseRef = `origin/${branchName}`;
    core.info(`[generate_git_patch] fetch succeeded, baseRef="${baseRef}"`);
    return baseRef;
  } catch (fetchErr) {
    // Branch doesn't exist on origin yet (new branch) – fall back to merge-base
    core.warning(`[generate_git_patch] fetch of origin/${branchName} failed (${getErrorMessage(fetchErr)}); falling back to merge-base with "${defaultBranch}"`);
  }

  execGitSync(["fetch", "origin", defaultBranch], { cwd });
  const baseRef = execGitSync(["merge-base", `origin/${defaultBranch}`, branchName], { cwd }).trim();
  core.info(`[generate_git_patch] merge-base baseRef="${baseRef}"`);
  return baseRef;
}

/**
 * Writes a patch file for the range base..tip and returns whether it succeeded.
 * @param {string} base - commit-ish for the base (exclusive)
 * @param {string} tip  - commit-ish for the tip (inclusive)
 * @param {string} cwd
 * @returns {boolean} true if the patch was written with content
 */
function writePatch(base, tip, cwd) {
  const commitCount = parseInt(execGitSync(["rev-list", "--count", `${base}..${tip}`], { cwd }).trim(), 10);
  core.info(`[generate_git_patch] ${commitCount} commit(s) between ${base} and ${tip}`);

  if (commitCount === 0) {
    return false;
  }

  const patchContent = execGitSync(["format-patch", `${base}..${tip}`, "--stdout"], { cwd });
  if (!patchContent || !patchContent.trim()) {
    core.warning(`[generate_git_patch] format-patch produced empty output for ${base}..${tip}`);
    return false;
  }

  fs.writeFileSync(PATCH_PATH, patchContent, "utf8");
  core.info(`[generate_git_patch] patch written: ${patchContent.split("\n").length} lines, ${Math.ceil(Buffer.byteLength(patchContent, "utf8") / 1024)} KB`);
  return true;
}

/**
 * Strategy 1: generate a patch from the known remote state of branchName to
 * its local tip, capturing only commits not yet on origin.
 * @param {string} branchName
 * @param {string} defaultBranch
 * @param {string} cwd
 * @returns {boolean} true if a patch was written
 */
function tryBranchStrategy(branchName, defaultBranch, cwd) {
  core.info(`[generate_git_patch] Strategy 1: branch-based patch for "${branchName}"`);
  try {
    execGitSync(["show-ref", "--verify", "--quiet", `refs/heads/${branchName}`], { cwd });
  } catch (err) {
    core.info(`[generate_git_patch] local branch "${branchName}" not found: ${getErrorMessage(err)}`);
    return false;
  }

  const baseRef = resolveBaseRef(branchName, defaultBranch, cwd);
  return writePatch(baseRef, branchName, cwd);
}

/**
 * Strategy 2: generate a patch from GITHUB_SHA to the current HEAD, capturing
 * commits made by the agent after checkout.
 * @param {string|undefined} githubSha
 * @param {string} cwd
 * @returns {{ generated: boolean, errorMessage: string|null }}
 */
function tryHeadStrategy(githubSha, cwd) {
  const currentHead = execGitSync(["rev-parse", "HEAD"], { cwd }).trim();
  core.info(`[generate_git_patch] Strategy 2: HEAD="${currentHead}" GITHUB_SHA="${githubSha || ""}"`);

  if (!githubSha) {
    const msg = "GITHUB_SHA environment variable is not set";
    core.warning(`[generate_git_patch] ${msg}`);
    return { generated: false, errorMessage: msg };
  }

  if (currentHead === githubSha) {
    core.info(`[generate_git_patch] HEAD matches GITHUB_SHA – no new commits`);
    return { generated: false, errorMessage: null };
  }

  try {
    execGitSync(["merge-base", "--is-ancestor", githubSha, "HEAD"], { cwd });
  } catch {
    core.warning(`[generate_git_patch] GITHUB_SHA is not an ancestor of HEAD – repository state has diverged`);
    return { generated: false, errorMessage: null };
  }

  const generated = writePatch(githubSha, "HEAD", cwd);
  return { generated, errorMessage: null };
}

/**
 * Generates a git patch file for the current changes.
 * @param {string} branchName - The branch name to generate patch for
 * @returns {Object} Object with patch info or error
 */
function generateGitPatch(branchName) {
  const cwd = process.env.GITHUB_WORKSPACE || process.cwd();
  const defaultBranch = process.env.DEFAULT_BRANCH || getBaseBranch();
  const githubSha = process.env.GITHUB_SHA;

  core.info(`[generate_git_patch] branchName="${branchName || ""}" GITHUB_SHA="${githubSha || ""}" defaultBranch="${defaultBranch}"`);

  const patchDir = path.dirname(PATCH_PATH);
  if (!fs.existsSync(patchDir)) {
    fs.mkdirSync(patchDir, { recursive: true });
  }

  let patchGenerated = false;
  let errorMessage = null;

  try {
    if (branchName) {
      patchGenerated = tryBranchStrategy(branchName, defaultBranch, cwd);
    } else {
      core.info(`[generate_git_patch] Strategy 1: skipped (no branchName)`);
    }

    if (!patchGenerated) {
      const result = tryHeadStrategy(githubSha, cwd);
      patchGenerated = result.generated;
      errorMessage = result.errorMessage;
    }
  } catch (error) {
    errorMessage = `Failed to generate patch: ${getErrorMessage(error)}`;
    core.warning(`[generate_git_patch] ${errorMessage}`);
  }

  if (patchGenerated && fs.existsSync(PATCH_PATH)) {
    const patchContent = fs.readFileSync(PATCH_PATH, "utf8");
    const patchSize = Buffer.byteLength(patchContent, "utf8");
    const patchLines = patchContent.split("\n").length;

    if (!patchContent.trim()) {
      return {
        success: false,
        error: "No changes to commit - patch is empty",
        patchPath: PATCH_PATH,
        patchSize: 0,
        patchLines: 0,
      };
    }

    return {
      success: true,
      patchPath: PATCH_PATH,
      patchSize: patchSize,
      patchLines: patchLines,
    };
  }

  return {
    success: false,
    error: errorMessage || "No changes to commit - no commits found",
    patchPath: PATCH_PATH,
  };
}

module.exports = {
  generateGitPatch,
};
