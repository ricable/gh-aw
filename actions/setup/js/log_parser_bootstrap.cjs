// @ts-check
/// <reference types="@actions/github-script" />

const { generatePlainTextSummary, generateCopilotCliStyleSummary, wrapAgentLogInSection, formatSafeOutputsPreview } = require("./log_parser_shared.cjs");
const { getErrorMessage } = require("./error_helpers.cjs");

/**
 * Add failure diagnostics to step summary when agent produces no logs
 * @param {string} message - Diagnostic message to display
 */
function addFailureDiagnostics(message) {
  const fs = require("fs");

  // Start building diagnostic markdown
  let diagnosticMarkdown = `## ⚠️ Agent Execution Diagnostics\n\n`;
  diagnosticMarkdown += `**Issue:** ${message}\n\n`;

  // Check for agent-stdio.log which captures stdout/stderr from the agent execution
  const stdioLogPath = "/tmp/gh-aw/agent-stdio.log";
  if (fs.existsSync(stdioLogPath)) {
    try {
      const stdioContent = fs.readFileSync(stdioLogPath, "utf8").trim();
      if (stdioContent) {
        // Extract last 50 lines for brevity (typical error messages are at the end)
        const lines = stdioContent.split("\n");
        const relevantLines = lines.slice(-50);
        const truncated = lines.length > 50;

        diagnosticMarkdown += `### Agent Execution Output\n\n`;
        if (truncated) {
          diagnosticMarkdown += `<details>\n<summary>Last 50 lines of agent output (${lines.length} total lines)</summary>\n\n`;
        }
        diagnosticMarkdown += "```\n";
        diagnosticMarkdown += relevantLines.join("\n");
        diagnosticMarkdown += "\n```\n";
        if (truncated) {
          diagnosticMarkdown += `\n</details>\n`;
        }
        diagnosticMarkdown += "\n";

        // Look for common error patterns
        const errorPatterns = [
          { pattern: /error|fail|exception/i, label: "Error indicators found in output" },
          { pattern: /command not found/i, label: "Command not found - possible installation issue" },
          { pattern: /permission denied/i, label: "Permission issue detected" },
          { pattern: /timeout|timed out/i, label: "Timeout detected" },
          { pattern: /killed/i, label: "Process was killed" },
        ];

        const detectedIssues = [];
        for (const { pattern, label } of errorPatterns) {
          if (pattern.test(stdioContent)) {
            detectedIssues.push(label);
          }
        }

        if (detectedIssues.length > 0) {
          diagnosticMarkdown += `### Detected Issues\n\n`;
          for (const issue of detectedIssues) {
            diagnosticMarkdown += `- ⚠️ ${issue}\n`;
          }
          diagnosticMarkdown += "\n";
        }
      }
    } catch (error) {
      core.warning(`Failed to read agent stdio log: ${getErrorMessage(error)}`);
    }
  }

  // Add troubleshooting guidance
  diagnosticMarkdown += `### Troubleshooting Steps\n\n`;
  diagnosticMarkdown += `1. **Check the full workflow logs** for the agent execution step\n`;
  diagnosticMarkdown += `2. **Review agent artifacts** if available (uploaded as workflow artifacts)\n`;
  diagnosticMarkdown += `3. **Check for resource constraints** (memory, timeout, disk space)\n`;
  diagnosticMarkdown += `4. **Verify agent installation** completed successfully in earlier steps\n`;
  diagnosticMarkdown += `5. **Review recent workflow changes** that might affect agent execution\n\n`;

  diagnosticMarkdown += `> 💡 **Tip:** Download the \`agent-artifacts\` artifact from this workflow run for detailed logs.\n`;

  // Write to step summary
  core.summary.addRaw(diagnosticMarkdown);
  core.summary.write();

  // Also log to console
  core.warning(`Agent execution diagnostics: ${message}`);
}

/**
 * Bootstrap helper for log parser entry points.
 * Handles common logic for environment variable lookup, file existence checks,
 * content reading (file or directory), and summary emission.
 *
 * @param {Object} options - Configuration options
 * @param {function(string): string|{markdown: string, mcpFailures?: string[], maxTurnsHit?: boolean, logEntries?: Array}} options.parseLog - Parser function that takes log content and returns markdown or result object
 * @param {string} options.parserName - Name of the parser (e.g., "Codex", "Claude", "Copilot")
 * @param {boolean} [options.supportsDirectories=false] - Whether the parser supports reading from directories
 * @returns {Promise<void>}
 */
async function runLogParser(options) {
  const fs = require("fs");
  const path = require("path");
  const { parseLog, parserName, supportsDirectories = false } = options;

  try {
    const logPath = process.env.GH_AW_AGENT_OUTPUT;
    if (!logPath) {
      core.info("No agent log file specified");
      addFailureDiagnostics("No GH_AW_AGENT_OUTPUT environment variable set");
      return;
    }

    if (!fs.existsSync(logPath)) {
      core.info(`Log path not found: ${logPath}`);
      addFailureDiagnostics(`Agent log directory not found: ${logPath}. This indicates the agent may have failed to start.`);
      return;
    }

    let content = "";

    // Check if logPath is a directory or a file
    const stat = fs.statSync(logPath);
    if (stat.isDirectory()) {
      if (!supportsDirectories) {
        core.info(`Log path is a directory but ${parserName} parser does not support directories: ${logPath}`);
        return;
      }

      // For Copilot, check if conversation.md exists (generated by --share flag)
      // If it exists, use it as the primary source for the step summary
      if (parserName === "Copilot") {
        const conversationMdPath = path.join(logPath, "conversation.md");
        if (fs.existsSync(conversationMdPath)) {
          core.info(`Found conversation.md generated by --share flag, using it for step summary preview`);
          content = fs.readFileSync(conversationMdPath, "utf8");

          // Transform markdown to increase header levels by 1
          // This adjusts the conversation.md headers (# to ##, etc.) for better display in step summary
          const { increaseHeaderLevel } = require("./markdown_transformer.cjs");
          const transformedContent = increaseHeaderLevel(content);

          // Mark this content as already markdown formatted
          // We'll need to adjust the parser to handle this
          const result = {
            markdown: transformedContent,
            isPreformatted: true,
            logEntries: [],
          };

          // Write to step summary directly
          if (result.markdown) {
            core.summary.addRaw(result.markdown);
            await core.summary.write();
            core.info(`Wrote conversation markdown to step summary (${Buffer.byteLength(result.markdown, "utf8")} bytes)`);
          }
          return;
        }
      }

      // Read all log files from the directory and concatenate them
      const files = fs.readdirSync(logPath);
      const logFiles = files.filter(file => file.endsWith(".log") || file.endsWith(".txt"));

      if (logFiles.length === 0) {
        core.info(`No log files found in directory: ${logPath}`);
        addFailureDiagnostics(`No log files found in ${logPath}. The agent did not produce any output logs.`);
        return;
      }

      // Sort log files by name to ensure consistent ordering
      logFiles.sort();

      // Concatenate all log files
      for (const file of logFiles) {
        const filePath = path.join(logPath, file);
        const fileContent = fs.readFileSync(filePath, "utf8");

        // Add a newline before this file if the previous content doesn't end with one
        if (content.length > 0 && !content.endsWith("\n")) {
          content += "\n";
        }

        content += fileContent;
      }
    } else {
      // Read the single log file
      content = fs.readFileSync(logPath, "utf8");
    }

    const result = parseLog(content);

    // Handle result that may be a simple string or an object with metadata
    let markdown = "";
    let mcpFailures = [];
    let maxTurnsHit = false;
    let logEntries = null;

    if (typeof result === "string") {
      markdown = result;
    } else if (result && typeof result === "object") {
      markdown = result.markdown || "";
      mcpFailures = result.mcpFailures || [];
      maxTurnsHit = result.maxTurnsHit || false;
      logEntries = result.logEntries || null;
    }

    if (markdown) {
      // Read safe outputs file if available
      let safeOutputsContent = "";
      const safeOutputsPath = process.env.GH_AW_SAFE_OUTPUTS;
      if (safeOutputsPath && fs.existsSync(safeOutputsPath)) {
        try {
          safeOutputsContent = fs.readFileSync(safeOutputsPath, "utf8");
        } catch (error) {
          core.warning(`Failed to read safe outputs file: ${getErrorMessage(error)}`);
        }
      }

      // Generate lightweight plain text summary for core.info and Copilot CLI style for step summary
      if (logEntries && Array.isArray(logEntries) && logEntries.length > 0) {
        // Extract model from init entry if available
        const initEntry = logEntries.find(entry => entry.type === "system" && entry.subtype === "init");
        const model = initEntry?.model || null;

        const plainTextSummary = generatePlainTextSummary(logEntries, {
          model,
          parserName,
        });
        core.info(plainTextSummary);

        // Add safe outputs preview to core.info
        if (safeOutputsContent) {
          const safeOutputsPlainText = formatSafeOutputsPreview(safeOutputsContent, { isPlainText: true });
          if (safeOutputsPlainText) {
            core.info(safeOutputsPlainText);
          }
        }

        // Generate Copilot CLI style markdown for step summary
        const copilotCliStyleMarkdown = generateCopilotCliStyleSummary(logEntries, {
          model,
          parserName,
        });

        // Wrap the agent log in a details/summary section (open by default)
        const wrappedAgentLog = wrapAgentLogInSection(copilotCliStyleMarkdown, {
          parserName,
          open: true,
        });

        // Add safe outputs preview to step summary
        let fullMarkdown = wrappedAgentLog;
        if (safeOutputsContent) {
          const safeOutputsMarkdown = formatSafeOutputsPreview(safeOutputsContent, { isPlainText: false });
          if (safeOutputsMarkdown) {
            fullMarkdown += "\n" + safeOutputsMarkdown;
          }
        }

        core.summary.addRaw(fullMarkdown).write();
      } else {
        // Fallback: just log success message for parsers without log entries
        core.info(`${parserName} log parsed successfully`);

        // Add safe outputs preview to core.info (fallback path)
        if (safeOutputsContent) {
          const safeOutputsPlainText = formatSafeOutputsPreview(safeOutputsContent, { isPlainText: true });
          if (safeOutputsPlainText) {
            core.info(safeOutputsPlainText);
          }
        }

        // Wrap the original markdown in a details/summary section (open by default)
        const wrappedAgentLog = wrapAgentLogInSection(markdown, {
          parserName,
          open: true,
        });

        // Write wrapped markdown to step summary if available
        let fullMarkdown = wrappedAgentLog;
        if (safeOutputsContent) {
          const safeOutputsMarkdown = formatSafeOutputsPreview(safeOutputsContent, { isPlainText: false });
          if (safeOutputsMarkdown) {
            fullMarkdown += "\n" + safeOutputsMarkdown;
          }
        }
        core.summary.addRaw(fullMarkdown).write();
      }
    } else {
      core.error(`Failed to parse ${parserName} log`);
    }

    // Handle MCP server failures if present
    if (mcpFailures && mcpFailures.length > 0) {
      const failedServers = mcpFailures.join(", ");
      core.setFailed(`MCP server(s) failed to launch: ${failedServers}`);
    }

    // Handle max-turns limit if hit
    if (maxTurnsHit) {
      core.setFailed(`Agent execution stopped: max-turns limit reached. The agent did not complete its task successfully.`);
    }
  } catch (error) {
    core.setFailed(error instanceof Error ? error : String(error));
  }
}

// Export for testing and usage
if (typeof module !== "undefined" && module.exports) {
  module.exports = {
    runLogParser,
  };
}
