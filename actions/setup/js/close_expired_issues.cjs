// @ts-check
// <reference types="@actions/github-script" />

const { executeExpiredEntityCleanup } = require("./expired_entity_main_flow.cjs");
const { generateExpiredEntityFooter } = require("./generate_footer.cjs");
const { sanitizeContent } = require("./sanitize_content.cjs");
const { getWorkflowMetadata } = require("./workflow_metadata_helpers.cjs");
const { createExpiredEntityProcessor } = require("./expired_entity_handlers.cjs");

/**
 * Add comment to a GitHub Issue using REST API
 * @param {any} github - GitHub REST instance
 * @param {string} owner - Repository owner
 * @param {string} repo - Repository name
 * @param {number} issueNumber - Issue number
 * @param {string} message - Comment body
 * @returns {Promise<any>} Comment details
 */
async function addIssueComment(github, owner, repo, issueNumber, message) {
  const result = await github.rest.issues.createComment({
    owner: owner,
    repo: repo,
    issue_number: issueNumber,
    body: sanitizeContent(message),
  });

  return result.data;
}

/**
 * Close a GitHub Issue using REST API
 * @param {any} github - GitHub REST instance
 * @param {string} owner - Repository owner
 * @param {string} repo - Repository name
 * @param {number} issueNumber - Issue number
 * @returns {Promise<any>} Issue details
 */
async function closeIssue(github, owner, repo, issueNumber) {
  const result = await github.rest.issues.update({
    owner: owner,
    repo: repo,
    issue_number: issueNumber,
    state: "closed",
    state_reason: "not_planned",
  });

  return result.data;
}

async function main() {
  const owner = context.repo.owner;
  const repo = context.repo.repo;

  // Get workflow metadata for footer
  const { workflowName, workflowId, runUrl } = getWorkflowMetadata(owner, repo);

  // Create processor using shared handler
  const processEntity = createExpiredEntityProcessor(workflowName, runUrl, workflowId, {
    entityType: "issue",
    addComment: async (issue, message) => {
      await addIssueComment(github, owner, repo, issue.number, message);
    },
    closeEntity: async issue => {
      await closeIssue(github, owner, repo, issue.number);
    },
    buildClosingMessage: (issue, workflowName, runUrl, workflowId) => {
      return `This issue was automatically closed because it expired on ${issue.expirationDate.toISOString()}.` + generateExpiredEntityFooter(workflowName, runUrl, workflowId);
    },
  });

  await executeExpiredEntityCleanup(github, owner, repo, {
    entityType: "issues",
    graphqlField: "issues",
    resultKey: "issues",
    entityLabel: "Issue",
    summaryHeading: "Expired Issues Cleanup",
    processEntity,
  });
}

module.exports = { main };
