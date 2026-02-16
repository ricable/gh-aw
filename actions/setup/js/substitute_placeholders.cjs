const fs = require("fs");
const { getErrorMessage } = require("./error_helpers.cjs");
const { validateAndNormalizePath } = require("./path_helpers.cjs");

const substitutePlaceholders = async ({ file, substitutions }) => {
  if (!file) throw new Error("file parameter is required");
  if (!substitutions || "object" != typeof substitutions) throw new Error("substitutions parameter must be an object");

  if (typeof core !== "undefined") {
    core.info(`[substitutePlaceholders] Starting placeholder substitution`);
    core.info(`[substitutePlaceholders] File (raw): ${file}`);
    core.info(`[substitutePlaceholders] Substitution count: ${Object.keys(substitutions).length}`);
  }

  // Validate and normalize the file path for security
  const validatedPath = validateAndNormalizePath(file, "file path");
  if (typeof core !== "undefined") {
    core.info(`[substitutePlaceholders] Validated file path: ${validatedPath}`);
  }

  let content;
  try {
    if (typeof core !== "undefined") {
      core.info(`[substitutePlaceholders] Reading file...`);
    }
    content = fs.readFileSync(validatedPath, "utf8");
    if (typeof core !== "undefined") {
      core.info(`[substitutePlaceholders] File size: ${content.length} characters`);
    }
  } catch (error) {
    const errorMessage = getErrorMessage(error);
    if (typeof core !== "undefined") {
      core.warning(`[substitutePlaceholders] Failed to read file: ${errorMessage}`);
    }
    throw new Error(`Failed to read file ${validatedPath}: ${errorMessage}`);
  }

  for (const [key, value] of Object.entries(substitutions)) {
    const placeholder = `__${key}__`;
    // Convert undefined/null to empty string to avoid leaving "undefined" or "null" in the output
    const safeValue = value === undefined || value === null ? "" : value;
    const occurrences = (content.match(new RegExp(placeholder.replace(/[.*+?^${}()|[\]\\]/g, "\\$&"), "g")) || []).length;

    if (occurrences > 0) {
      if (typeof core !== "undefined") {
        core.info(`[substitutePlaceholders] Replacing placeholder: ${placeholder} (${occurrences} occurrence(s))`);
        core.info(`[substitutePlaceholders]   Value: ${String(safeValue).substring(0, 100)}${String(safeValue).length > 100 ? "..." : ""}`);
      }
      content = content.split(placeholder).join(safeValue);
    } else {
      if (typeof core !== "undefined") {
        core.info(`[substitutePlaceholders] Placeholder not found: ${placeholder} (unused)`);
      }
    }
  }

  try {
    if (typeof core !== "undefined") {
      core.info(`[substitutePlaceholders] Writing updated content back to file...`);
    }
    fs.writeFileSync(validatedPath, content, "utf8");
    if (typeof core !== "undefined") {
      core.info(`[substitutePlaceholders] ✓ Successfully substituted ${Object.keys(substitutions).length} placeholder(s)`);
    }
  } catch (error) {
    const errorMessage = getErrorMessage(error);
    if (typeof core !== "undefined") {
      core.warning(`[substitutePlaceholders] Failed to write file: ${errorMessage}`);
    }
    throw new Error(`Failed to write file ${validatedPath}: ${errorMessage}`);
  }
  return `Successfully substituted ${Object.keys(substitutions).length} placeholder(s) in ${validatedPath}`;
};
module.exports = substitutePlaceholders;
