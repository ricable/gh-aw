// @ts-check

/**
 * GitHub Actions Core Object Shim
 *
 * This module provides a lightweight shim of the @actions/core object for use
 * in Node.js environments where @actions/core is not available (e.g., MCP servers,
 * safe-inputs, safe-outputs running outside GitHub Actions).
 *
 * When running in GitHub Actions context with @actions/core available, this shim
 * is not needed. However, when the same code needs to run in a Node.js environment,
 * this shim provides compatible logging methods that fall back to console output.
 *
 * Usage:
 *   const core = require("./core_shim.cjs");
 *   core.info("Information message");
 *   core.warning("Warning message");
 *   core.error("Error message");
 *   core.debug("Debug message");
 *   core.setFailed("Fatal error message");
 */

/**
 * Core shim object that provides GitHub Actions core-compatible logging
 * methods that work in both GitHub Actions and Node.js environments.
 */
const coreShim = {
  /**
   * Write an informational message
   * @param {string} message - The message to log
   */
  info(message) {
    console.log(`[INFO] ${message}`);
  },

  /**
   * Write a warning message
   * @param {string} message - The warning message
   */
  warning(message) {
    console.warn(`[WARNING] ${message}`);
  },

  /**
   * Write an error message
   * @param {string} message - The error message
   */
  error(message) {
    console.error(`[ERROR] ${message}`);
  },

  /**
   * Write a debug message
   * @param {string} message - The debug message
   */
  debug(message) {
    // Only log debug messages if DEBUG environment variable is set
    if (process.env.DEBUG) {
      console.log(`[DEBUG] ${message}`);
    }
  },

  /**
   * Mark the action as failed with an error message
   * @param {string} message - The failure message
   */
  setFailed(message) {
    console.error(`[FAILED] ${message}`);
    // Note: We don't call process.exit() here because the caller should handle
    // the error appropriately. This is just for logging the failure.
  },
};

module.exports = coreShim;
