// @ts-check
/// <reference types="@actions/github-script" />

const { createEngineLogParser, formatToolCallAsDetails, estimateTokens, formatDuration } = require("./log_parser_shared.cjs");
const { unfenceMarkdown } = require("./markdown_unfencing.cjs");

const main = createEngineLogParser({
  parserName: "Gemini",
  parseFunction: parseGeminiLog,
  supportsDirectories: false,
});

/**
 * Parse Gemini CLI JSONL log output and format as markdown.
 *
 * Gemini outputs one JSON object per line with the following entry types:
 *   - init:        { type, timestamp, session_id, model }
 *   - message:     { type, timestamp, role, content, delta? }
 *                  All assistant messages have delta=true and are streaming chunks
 *                  that must be concatenated to form complete turns.
 *   - tool_use:    { type, timestamp, tool_name, tool_id, parameters }
 *   - tool_result: { type, timestamp, tool_id, status, output }
 *   - result:      { type, timestamp, status, stats: { total_tokens, input_tokens,
 *                    output_tokens, cached, duration_ms, tool_calls } }
 *
 * Non-JSON lines (e.g. [INFO], [WARN] from the runner wrapper) are silently skipped.
 *
 * @param {string} logContent - The raw log content to parse
 * @returns {{markdown: string, logEntries: Array, mcpFailures: Array<string>, maxTurnsHit: boolean}} Parsed log data
 */
function parseGeminiLog(logContent) {
  if (!logContent) {
    return {
      markdown: "## 🤖 Gemini\n\nNo log content provided.\n\n",
      logEntries: [],
      mcpFailures: [],
      maxTurnsHit: false,
    };
  }

  // Parse all JSONL entries, skipping non-JSON lines
  /** @type {Array<any>} */
  const entries = [];
  for (const line of logContent.split("\n")) {
    const trimmed = line.trim();
    if (!trimmed.startsWith("{")) continue;
    try {
      entries.push(JSON.parse(trimmed));
    } catch (_e) {
      // Skip invalid JSON lines
    }
  }

  if (entries.length === 0) {
    return {
      markdown: "## 🤖 Gemini\n\nNo log content provided.\n\n",
      logEntries: [],
      mcpFailures: [],
      maxTurnsHit: false,
    };
  }

  // Index tool results by tool_id for O(1) lookup when rendering tool calls
  /** @type {Map<string, any>} */
  const toolResultMap = new Map();
  for (const entry of entries) {
    if (entry.type === "tool_result" && entry.tool_id) {
      toolResultMap.set(entry.tool_id, entry);
    }
  }

  // Find the final result entry that carries aggregate stats
  const resultEntry = entries.find(e => e.type === "result");

  let markdown = "";
  markdown += "## 🤖 Reasoning\n\n";

  // Walk entries in order.  Consecutive assistant delta chunks are accumulated
  // into a single thought, which is flushed as soon as a non-message entry
  // (or a non-assistant message) is encountered.
  let currentThought = "";
  /** @type {Array<string>} */
  const commandSummary = [];

  /**
   * Flush the accumulated assistant thought to markdown.
   */
  function flushThought() {
    if (!currentThought.trim()) {
      currentThought = "";
      return;
    }
    const text = unfenceMarkdown(currentThought.trim());
    if (text) {
      markdown += text + "\n\n";
    }
    currentThought = "";
  }

  for (const entry of entries) {
    if (entry.type === "message" && entry.role === "assistant" && entry.delta) {
      // Accumulate streaming chunks from the assistant
      currentThought += entry.content || "";
    } else if (entry.type === "tool_use") {
      // Flush any pending assistant thought before rendering the tool call
      flushThought();

      const toolName = entry.tool_name || "unknown";
      const toolResult = toolResultMap.get(entry.tool_id);
      const statusIcon = toolResult ? (toolResult.status === "success" ? "✅" : "❌") : "❓";

      const params = entry.parameters ? JSON.stringify(entry.parameters, null, 2) : "";
      const output = toolResult ? String(toolResult.output || "") : "";

      markdown += formatGeminiToolCall(toolName, params, output, statusIcon);
      commandSummary.push(`* ${statusIcon} \`${toolName}(...)\``);
    }
    // init, message(user), tool_result, and result entries are handled
    // either above or via the resultEntry / toolResultMap lookups.
  }

  // Flush any remaining accumulated thought
  flushThought();

  // Commands and Tools section
  markdown += "## 🤖 Commands and Tools\n\n";
  if (commandSummary.length > 0) {
    for (const cmd of commandSummary) {
      markdown += cmd + "\n";
    }
    markdown += "\n";
  } else {
    markdown += "No commands or tools used.\n";
  }

  // Information section
  markdown += "\n## 📊 Information\n\n";
  if (resultEntry && resultEntry.stats) {
    const stats = resultEntry.stats;

    if (stats.total_tokens > 0) {
      markdown += "**Token Usage:**\n";
      markdown += `- Total: ${stats.total_tokens.toLocaleString()}\n`;
      if (stats.input_tokens) markdown += `- Input: ${stats.input_tokens.toLocaleString()}\n`;
      if (stats.cached) markdown += `- Cached: ${stats.cached.toLocaleString()}\n`;
      if (stats.output_tokens) markdown += `- Output: ${stats.output_tokens.toLocaleString()}\n`;
      markdown += "\n";
    }

    if (stats.duration_ms) {
      markdown += `**Duration:** ${formatDuration(stats.duration_ms)}\n\n`;
    }

    if (stats.tool_calls) {
      markdown += `**Tool Calls:** ${stats.tool_calls}\n\n`;
    }
  }

  return {
    markdown,
    logEntries: [],
    mcpFailures: [],
    maxTurnsHit: false,
  };
}

/**
 * Format a Gemini tool call with HTML details block.
 * Uses the shared formatToolCallAsDetails helper for consistent rendering.
 *
 * @param {string} toolName - The tool name (e.g., "list_pull_requests")
 * @param {string} params - The parameters as a JSON string
 * @param {string} output - The tool output string
 * @param {string} statusIcon - The status icon (✅, ❌, or ❓)
 * @returns {string} Formatted HTML details string
 */
function formatGeminiToolCall(toolName, params, output, statusIcon) {
  const totalTokens = estimateTokens(params) + estimateTokens(output);
  const metadata = totalTokens > 0 ? `<code>~${totalTokens}t</code>` : "";
  const summary = `<code>${toolName}</code>`;

  /** @type {Array<{label: string, content: string, language?: string}>} */
  const sections = [];
  if (params && params.trim()) {
    sections.push({ label: "Parameters", content: params, language: "json" });
  }
  if (output && output.trim()) {
    sections.push({ label: "Output", content: output });
  }

  return formatToolCallAsDetails({ summary, statusIcon, metadata, sections });
}

// Export for testing
if (typeof module !== "undefined" && module.exports) {
  module.exports = {
    main,
    parseGeminiLog,
    formatGeminiToolCall,
  };
}
