// @ts-check
/// <reference types="@actions/github-script" />

/**
 * File Helper Functions
 *
 * This module provides helper functions for file operations,
 * particularly for listing directories and checking file existence
 * with helpful error messages.
 */

const fs = require("fs");
const path = require("path");
const { getErrorMessage } = require("./error_helpers.cjs");
const { safeJoin } = require("./path_helpers.cjs");

/**
 * List all files recursively in a directory
 * @param {string} dirPath - The directory path to list
 * @param {string} [relativeTo] - Optional base path to show relative paths
 * @returns {string[]} Array of file paths
 */
function listFilesRecursively(dirPath, relativeTo) {
  const files = [];
  try {
    core.info(`[listFilesRecursively] Listing files in: ${dirPath}`);
    if (!fs.existsSync(dirPath)) {
      core.info(`[listFilesRecursively] Directory does not exist: ${dirPath}`);
      return files;
    }
    const entries = fs.readdirSync(dirPath, { withFileTypes: true });
    core.info(`[listFilesRecursively] Found ${entries.length} entries in ${dirPath}`);
    for (const entry of entries) {
      const fullPath = safeJoin(dirPath, entry.name);
      if (entry.isDirectory()) {
        core.info(`[listFilesRecursively] Recursing into directory: ${entry.name}`);
        files.push(...listFilesRecursively(fullPath, relativeTo));
      } else {
        const displayPath = relativeTo ? path.relative(relativeTo, fullPath) : fullPath;
        core.info(`[listFilesRecursively] Found file: ${displayPath}`);
        files.push(displayPath);
      }
    }
    core.info(`[listFilesRecursively] Total files found: ${files.length}`);
  } catch (error) {
    core.warning("Failed to list files in " + dirPath + ": " + getErrorMessage(error));
  }
  return files;
}

/**
 * Check if file exists and provide helpful error message with directory listing
 * @param {string} filePath - The file path to check
 * @param {string} artifactDir - The artifact directory to list if file not found
 * @param {string} fileDescription - Description of the file (e.g., "Prompt file", "Agent output file")
 * @param {boolean} required - Whether the file is required
 * @returns {boolean} True if file exists (or not required), false otherwise
 */
function checkFileExists(filePath, artifactDir, fileDescription, required) {
  core.info(`[checkFileExists] Checking ${fileDescription}: ${filePath}`);
  core.info(`[checkFileExists] Required: ${required}`);

  if (fs.existsSync(filePath)) {
    try {
      const stats = fs.statSync(filePath);
      const fileInfo = filePath + " (" + stats.size + " bytes)";
      core.info(`[checkFileExists] ✓ ${fileDescription} found: ${fileInfo}`);
      core.info(fileDescription + " found: " + fileInfo);
      return true;
    } catch (error) {
      core.warning("Failed to stat " + fileDescription.toLowerCase() + ": " + getErrorMessage(error));
      return false;
    }
  } else {
    if (required) {
      core.warning(`[checkFileExists] ❌ ${fileDescription} not found at: ${filePath}`);
      core.warning("❌ " + fileDescription + " not found at: " + filePath);
      // List all files in artifact directory for debugging
      core.info(`[checkFileExists] Listing artifact directory for debugging: ${artifactDir}`);
      core.info("📁 Listing all files in artifact directory: " + artifactDir);
      const files = listFilesRecursively(artifactDir, artifactDir);
      if (files.length === 0) {
        core.warning("  No files found in " + artifactDir);
      } else {
        core.info("  Found " + files.length + " file(s):");
        files.forEach(file => core.info("    - " + file));
      }
      core.setFailed("❌ " + fileDescription + " not found at: " + filePath);
      return false;
    } else {
      core.info(`[checkFileExists] No ${fileDescription.toLowerCase()} found at: ${filePath} (optional)`);
      core.info("No " + fileDescription.toLowerCase() + " found at: " + filePath);
      return true;
    }
  }
}

module.exports = { listFilesRecursively, checkFileExists };
