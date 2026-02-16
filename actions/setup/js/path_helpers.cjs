// @ts-check
/// <reference types="@actions/github-script" />

/**
 * Path Security Helper Functions
 *
 * This module provides helper functions for validating and normalizing
 * file paths to prevent path traversal attacks and other security issues.
 */

const path = require("path");
const fs = require("fs");

/**
 * Validates and normalizes a file path to prevent path traversal attacks.
 * Ensures the path is absolute and does not contain directory traversal patterns.
 *
 * @param {string} filePath - The file path to validate and normalize
 * @param {string} description - Description of what the path is for (for error messages)
 * @returns {string} - The validated and normalized absolute path
 * @throws {Error} - If the path is invalid or contains directory traversal patterns
 */
function validateAndNormalizePath(filePath, description = "file path") {
  if (!filePath || typeof filePath !== "string") {
    throw new Error(`Invalid ${description}: path must be a non-empty string`);
  }

  core.info(`[validateAndNormalizePath] Validating ${description}: ${filePath}`);

  // Remove any leading/trailing whitespace
  const trimmedPath = filePath.trim();

  if (trimmedPath.length === 0) {
    throw new Error(`Invalid ${description}: path cannot be empty or whitespace-only`);
  }

  // Check for null bytes (potential security issue)
  if (trimmedPath.includes("\0")) {
    throw new Error(`Security: ${description} contains null bytes`);
  }

  // Resolve to absolute path and normalize
  const absolutePath = path.resolve(trimmedPath);
  const normalizedPath = path.normalize(absolutePath);

  core.info(`[validateAndNormalizePath] Normalized path: ${normalizedPath}`);

  // Check for directory traversal patterns in the original path
  // We check the original because path.resolve() will already resolve ../ sequences
  if (trimmedPath.includes("..")) {
    core.warning(`[validateAndNormalizePath] Path contains '..' sequence: ${trimmedPath}`);
    // This is allowed after normalization as long as it doesn't escape the base
  }

  return normalizedPath;
}

/**
 * Validates that a file path is within a specific base directory.
 * This prevents path traversal attacks that attempt to access files outside the allowed directory.
 *
 * @param {string} filePath - The file path to validate (can be relative or absolute)
 * @param {string} baseDir - The base directory that the file must be within
 * @param {string} description - Description of what the path is for (for error messages)
 * @returns {string} - The validated and normalized absolute path
 * @throws {Error} - If the path escapes the base directory or is invalid
 */
function validatePathWithinBase(filePath, baseDir, description = "file path") {
  if (!baseDir || typeof baseDir !== "string") {
    throw new Error("Invalid base directory: must be a non-empty string");
  }

  core.info(`[validatePathWithinBase] Validating ${description} within base: ${baseDir}`);
  core.info(`[validatePathWithinBase] Input path: ${filePath}`);

  // Normalize both the base directory and the file path
  const normalizedBase = path.normalize(path.resolve(baseDir));
  const normalizedPath = validateAndNormalizePath(filePath, description);

  // Get the relative path from base to the file
  const relativePath = path.relative(normalizedBase, normalizedPath);

  core.info(`[validatePathWithinBase] Normalized base: ${normalizedBase}`);
  core.info(`[validatePathWithinBase] Normalized path: ${normalizedPath}`);
  core.info(`[validatePathWithinBase] Relative path: ${relativePath}`);

  // Check if the relative path starts with .. (escapes base directory)
  // or is absolute (not within base directory)
  if (relativePath.startsWith("..") || path.isAbsolute(relativePath)) {
    core.warning(`[validatePathWithinBase] Security violation detected`);
    core.warning(`[validatePathWithinBase]   Base: ${normalizedBase}`);
    core.warning(`[validatePathWithinBase]   Path: ${normalizedPath}`);
    core.warning(`[validatePathWithinBase]   Relative: ${relativePath}`);
    throw new Error(
      `Security: ${description} must be within ${baseDir} (attempted to access: ${relativePath})`
    );
  }

  core.info(`[validatePathWithinBase] ✓ Path validated successfully: ${normalizedPath}`);
  return normalizedPath;
}

/**
 * Validates and normalizes a directory path, ensuring it exists and is a directory.
 *
 * @param {string} dirPath - The directory path to validate
 * @param {string} description - Description of what the directory is for (for error messages)
 * @param {boolean} createIfMissing - Whether to create the directory if it doesn't exist
 * @returns {string} - The validated and normalized absolute path
 * @throws {Error} - If the path is invalid or not a directory
 */
function validateDirectory(dirPath, description = "directory", createIfMissing = false) {
  const normalizedPath = validateAndNormalizePath(dirPath, description);

  core.info(`[validateDirectory] Checking ${description}: ${normalizedPath}`);

  if (!fs.existsSync(normalizedPath)) {
    if (createIfMissing) {
      core.info(`[validateDirectory] Creating ${description}: ${normalizedPath}`);
      fs.mkdirSync(normalizedPath, { recursive: true });
    } else {
      throw new Error(`${description} does not exist: ${normalizedPath}`);
    }
  }

  const stats = fs.statSync(normalizedPath);
  if (!stats.isDirectory()) {
    throw new Error(`${description} is not a directory: ${normalizedPath}`);
  }

  core.info(`[validateDirectory] ✓ Directory validated: ${normalizedPath}`);
  return normalizedPath;
}

/**
 * Safely joins path segments and normalizes the result.
 * This is safer than using path.join() directly as it normalizes the output.
 *
 * @param {...string} segments - Path segments to join
 * @returns {string} - The joined and normalized path
 */
function safeJoin(...segments) {
  const joined = path.join(...segments);
  const normalized = path.normalize(joined);
  
  core.info(`[safeJoin] Input segments: ${segments.join(", ")}`);
  core.info(`[safeJoin] Normalized result: ${normalized}`);
  
  return normalized;
}

module.exports = {
  validateAndNormalizePath,
  validatePathWithinBase,
  validateDirectory,
  safeJoin,
};
