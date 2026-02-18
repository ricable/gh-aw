// @ts-check
// <reference types="@actions/github-script" />

/**
 * Expired Entity Handlers
 *
 * This module provides reusable handlers for processing expired entities (issues, PRs, discussions).
 * It extracts the common comment + close + return record flow that was duplicated across
 * close_expired_issues.cjs, close_expired_pull_requests.cjs, and close_expired_discussions.cjs.
 */

/**
 * Configuration for entity-specific operations
 * @typedef {Object} EntityHandlerConfig
 * @property {string} entityType - Entity type for logging (e.g., "issue", "pull request", "discussion")
 * @property {(entity: any, message: string) => Promise<any>} addComment - Function to add a comment (receives entity and message)
 * @property {(entity: any) => Promise<any>} closeEntity - Function to close the entity (receives entity)
 * @property {(entity: any, workflowName: string, runUrl: string, workflowId: string) => string} buildClosingMessage - Function to build the closing message
 * @property {(entity: any) => Promise<{shouldSkip: boolean, reason?: string, shouldClose?: boolean}>} [preCheck] - Optional pre-check function (e.g., for duplicate detection)
 */

/**
 * Create a standard expired entity processor
 *
 * This function returns a processEntity function that can be passed to executeExpiredEntityCleanup.
 * It handles the common flow:
 * 1. Optional pre-check (e.g., checking for existing comments in discussions)
 * 2. Add closing comment
 * 3. Close entity
 * 4. Return status and record
 *
 * @param {string} workflowName - Workflow name for footer
 * @param {string} runUrl - Workflow run URL for footer
 * @param {string} workflowId - Workflow ID for footer
 * @param {EntityHandlerConfig} config - Entity-specific configuration
 * @returns {(entity: any) => Promise<{status: "closed" | "skipped", record: any}>}
 */
function createExpiredEntityProcessor(workflowName, runUrl, workflowId, config) {
  return async entity => {
    // Step 1: Optional pre-check (e.g., duplicate detection for discussions)
    if (config.preCheck) {
      const preCheckResult = await config.preCheck(entity);
      if (preCheckResult.shouldSkip) {
        if (preCheckResult.reason) {
          core.warning(`  ${preCheckResult.reason}`);
        }

        // If preCheck says to close without adding comment, do that
        if (preCheckResult.shouldClose) {
          core.info(`  Attempting to close ${config.entityType} #${entity.number} without adding another comment`);
          await config.closeEntity(entity);
          core.info(`  ✓ ${capitalize(config.entityType)} closed successfully`);
        }

        return {
          status: "skipped",
          record: {
            number: entity.number,
            url: entity.url,
            title: entity.title,
          },
        };
      }
    }

    // Step 2: Build closing message
    const closingMessage = config.buildClosingMessage(entity, workflowName, runUrl, workflowId);

    // Step 3: Add closing comment
    core.info(`  Adding closing comment to ${config.entityType} #${entity.number}`);
    await config.addComment(entity, closingMessage);
    core.info(`  ✓ Comment added successfully`);

    // Step 4: Close entity
    core.info(`  Closing ${config.entityType} #${entity.number}`);
    await config.closeEntity(entity);
    core.info(`  ✓ ${capitalize(config.entityType)} closed successfully`);

    // Step 5: Return status and record
    return {
      status: "closed",
      record: {
        number: entity.number,
        url: entity.url,
        title: entity.title,
      },
    };
  };
}

/**
 * Capitalize the first letter of a string
 * @param {string} str - String to capitalize
 * @returns {string}
 */
function capitalize(str) {
  return str.charAt(0).toUpperCase() + str.slice(1);
}

module.exports = {
  createExpiredEntityProcessor,
};
