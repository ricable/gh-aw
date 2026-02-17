// @ts-check
/// <reference types="@actions/github-script" />

const { createEngineLogParser, truncateString, estimateTokens, formatToolCallAsDetails } = require("./log_parser_shared.cjs");

const main = createEngineLogParser({
  parserName: "OpenClaw",
  parseFunction: parseOpenClawLog,
  supportsDirectories: false,
});

/**
 * Parse OpenClaw log content and format as markdown
 * OpenClaw's --json flag produces structured JSON output (one JSON object per line)
 * @param {string} logContent - The raw log content to parse
 * @returns {{markdown: string, logEntries: Array, mcpFailures: Array<string>, maxTurnsHit: boolean}} Parsed log data
 */
function parseOpenClawLog(logContent) {
  if (!logContent) {
    return {
      markdown: "## Agent Log Summary\n\nNo log content provided.\n\n",
      logEntries: [],
      mcpFailures: [],
      maxTurnsHit: false,
    };
  }

  const lines = logContent.split("\n");
  let markdown = "";
  const logEntries = [];
  const mcpFailures = [];
  let toolCallCount = 0;
  let totalTokens = 0;

  markdown += "## 🤖 Reasoning\n\n";

  // Parse each line as potential JSON
  for (const line of lines) {
    const trimmed = line.trim();
    if (!trimmed) continue;

    // Try to parse as JSON
    let entry;
    try {
      entry = JSON.parse(trimmed);
    } catch {
      // Not JSON - could be plain text output
      // Skip metadata/debug lines
      if (trimmed.startsWith("{") || trimmed.startsWith("[")) {
        continue; // Malformed JSON, skip
      }
      // Include non-JSON text in reasoning if substantive
      if (trimmed.length > 20 && !trimmed.match(/^\d{4}-\d{2}-\d{2}T/)) {
        markdown += `${trimmed}\n\n`;
      }
      continue;
    }

    // Process structured JSON entries
    if (!entry || typeof entry !== "object") continue;

    const entryType = entry.type || entry.event || "";

    switch (entryType) {
      case "message":
      case "thinking":
      case "reasoning": {
        const content = entry.content || entry.text || entry.message || "";
        if (content && content.length > 10) {
          markdown += `${truncateString(content, 500)}\n\n`;
          logEntries.push({
            type: "assistant",
            message: {
              content: [{ type: "text", text: content }],
            },
          });
        }
        break;
      }

      case "tool_call":
      case "tool_use": {
        toolCallCount++;
        const toolName = entry.name || entry.tool || entry.function || "unknown";
        const params = entry.input || entry.arguments || entry.params || {};
        const paramsStr = typeof params === "string" ? params : JSON.stringify(params, null, 2);

        markdown += formatOpenClawToolCall(toolName, paramsStr, "", "⏳");

        const toolUseId = `tool_${logEntries.length}`;
        logEntries.push({
          type: "assistant",
          message: {
            content: [
              {
                type: "tool_use",
                id: toolUseId,
                name: toolName,
                input: typeof params === "string" ? { params } : params,
              },
            ],
          },
        });
        break;
      }

      case "tool_result": {
        const result = entry.result || entry.output || entry.content || "";
        const resultStr = typeof result === "string" ? result : JSON.stringify(result, null, 2);
        const isError = entry.is_error || entry.error || false;
        const toolUseId = `tool_result_${logEntries.length}`;

        logEntries.push({
          type: "user",
          message: {
            content: [
              {
                type: "tool_result",
                tool_use_id: toolUseId,
                content: resultStr,
                is_error: isError,
              },
            ],
          },
        });
        break;
      }

      case "error": {
        const errorMsg = entry.message || entry.error || JSON.stringify(entry);
        markdown += `❌ **Error:** ${truncateString(errorMsg, 200)}\n\n`;
        break;
      }

      case "mcp_init":
      case "mcp_connected": {
        const serverName = entry.server || entry.name || "";
        if (serverName) {
          markdown += `✅ MCP Server connected: **${serverName}**\n\n`;
        }
        break;
      }

      case "mcp_error":
      case "mcp_failed": {
        const serverName = entry.server || entry.name || "";
        const error = entry.error || entry.message || "";
        if (serverName) {
          mcpFailures.push(serverName);
          markdown += `❌ MCP Server failed: **${serverName}** - ${error}\n\n`;
        }
        break;
      }

      case "usage":
      case "token_usage": {
        const tokens = entry.total_tokens || entry.tokens || 0;
        if (tokens > totalTokens) {
          totalTokens = tokens;
        }
        break;
      }

      default:
        // Skip unknown entry types
        break;
    }
  }

  // Add commands and tools section
  markdown += "## 🤖 Commands and Tools\n\n";
  if (toolCallCount > 0) {
    markdown += `**Tool Calls:** ${toolCallCount}\n\n`;
  } else {
    markdown += "No tool calls detected.\n\n";
  }

  // Add information section
  markdown += "## 📊 Information\n\n";
  if (totalTokens > 0) {
    markdown += `**Total Tokens Used:** ${totalTokens.toLocaleString()}\n\n`;
  }

  return {
    markdown,
    logEntries,
    mcpFailures,
    maxTurnsHit: false, // OpenClaw uses timeout, not max-turns
  };
}

/**
 * Format an OpenClaw tool call with HTML details
 * @param {string} toolName - The tool name
 * @param {string} params - The parameters as JSON string
 * @param {string} response - The response as JSON string
 * @param {string} statusIcon - The status icon
 * @returns {string} Formatted HTML details string
 */
function formatOpenClawToolCall(toolName, params, response, statusIcon) {
  const totalTokens = estimateTokens(params) + estimateTokens(response);

  let metadata = "";
  if (totalTokens > 0) {
    metadata = `<code>~${totalTokens}t</code>`;
  }

  const summary = `<code>${toolName}</code>`;

  const sections = [];

  if (params && params.trim()) {
    sections.push({
      label: "Parameters",
      content: params,
      language: "json",
    });
  }

  if (response && response.trim()) {
    sections.push({
      label: "Response",
      content: response,
      language: "json",
    });
  }

  return formatToolCallAsDetails({
    summary,
    statusIcon,
    metadata,
    sections,
  });
}

// Export for testing
if (typeof module !== "undefined" && module.exports) {
  module.exports = {
    main,
    parseOpenClawLog,
    formatOpenClawToolCall,
  };
}
