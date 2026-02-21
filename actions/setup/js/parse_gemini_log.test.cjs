import { describe, it, expect, beforeEach, vi } from "vitest";

describe("parse_gemini_log.cjs", () => {
  let mockCore;
  let parseGeminiLog;
  let formatGeminiToolCall;

  beforeEach(async () => {
    mockCore = {
      debug: vi.fn(),
      info: vi.fn(),
      warning: vi.fn(),
      error: vi.fn(),
      setFailed: vi.fn(),
      setOutput: vi.fn(),
      summary: {
        addRaw: vi.fn().mockReturnThis(),
        write: vi.fn().mockResolvedValue(undefined),
      },
    };
    global.core = mockCore;

    const module = await import("./parse_gemini_log.cjs?" + Date.now());
    parseGeminiLog = module.parseGeminiLog;
    formatGeminiToolCall = module.formatGeminiToolCall;
  });

  describe("parseGeminiLog function", () => {
    it("should return empty content message when log is null", () => {
      const result = parseGeminiLog(null);

      expect(result.markdown).toContain("No log content provided");
      expect(result.logEntries).toEqual([]);
      expect(result.mcpFailures).toEqual([]);
      expect(result.maxTurnsHit).toBe(false);
    });

    it("should return empty content message when log is empty string", () => {
      const result = parseGeminiLog("");

      expect(result.markdown).toContain("No log content provided");
    });

    it("should return empty content message when log has no JSON lines", () => {
      const logContent = `[INFO] Starting containers...
[WARN] Using --env-all
[SUCCESS] Containers started`;

      const result = parseGeminiLog(logContent);

      expect(result.markdown).toContain("No log content provided");
    });

    it("should parse assistant reasoning from delta messages", () => {
      const logContent = [
        JSON.stringify({ type: "init", timestamp: "2026-01-01T00:00:00Z", model: "auto-gemini-3", session_id: "abc123" }),
        JSON.stringify({ type: "message", role: "user", content: "Do a task", delta: false }),
        JSON.stringify({ type: "message", role: "assistant", content: "I will list", delta: true }),
        JSON.stringify({ type: "message", role: "assistant", content: " the pull requests.\n", delta: true }),
        JSON.stringify({ type: "result", timestamp: "2026-01-01T00:01:00Z", status: "success", stats: { total_tokens: 100, input_tokens: 80, output_tokens: 20, cached: 0, duration_ms: 5000, tool_calls: 0 } }),
      ].join("\n");

      const result = parseGeminiLog(logContent);

      expect(result.markdown).toContain("## 🤖 Reasoning");
      expect(result.markdown).toContain("I will list the pull requests.");
    });

    it("should concatenate consecutive delta messages into a single thought", () => {
      const logContent = [
        JSON.stringify({ type: "message", role: "assistant", content: "Part one", delta: true }),
        JSON.stringify({ type: "message", role: "assistant", content: " and part two.", delta: true }),
        JSON.stringify({ type: "result", status: "success", stats: { total_tokens: 50, input_tokens: 30, output_tokens: 20, cached: 0, duration_ms: 1000, tool_calls: 0 } }),
      ].join("\n");

      const result = parseGeminiLog(logContent);

      expect(result.markdown).toContain("Part one and part two.");
    });

    it("should parse tool_use and tool_result pairs", () => {
      const logContent = [
        JSON.stringify({ type: "message", role: "assistant", content: "I will list PRs.", delta: true }),
        JSON.stringify({ type: "tool_use", timestamp: "2026-01-01T00:00:01Z", tool_name: "list_pull_requests", tool_id: "lpr_001", parameters: { owner: "github", repo: "gh-aw" } }),
        JSON.stringify({ type: "tool_result", timestamp: "2026-01-01T00:00:02Z", tool_id: "lpr_001", status: "success", output: '{"items":[]}' }),
        JSON.stringify({ type: "result", status: "success", stats: { total_tokens: 200, input_tokens: 150, output_tokens: 50, cached: 0, duration_ms: 3000, tool_calls: 1 } }),
      ].join("\n");

      const result = parseGeminiLog(logContent);

      expect(result.markdown).toContain("list_pull_requests");
      expect(result.markdown).toContain("✅");
      expect(result.markdown).toContain("## 🤖 Commands and Tools");
      expect(result.markdown).toContain("* ✅ `list_pull_requests(...)`");
    });

    it("should mark tool_result with error status as failure", () => {
      const logContent = [
        JSON.stringify({ type: "tool_use", tool_name: "create_issue", tool_id: "ci_001", parameters: { title: "Test" } }),
        JSON.stringify({ type: "tool_result", tool_id: "ci_001", status: "error", output: "Permission denied" }),
        JSON.stringify({ type: "result", status: "success", stats: { total_tokens: 100, input_tokens: 80, output_tokens: 20, cached: 0, duration_ms: 1000, tool_calls: 1 } }),
      ].join("\n");

      const result = parseGeminiLog(logContent);

      expect(result.markdown).toContain("❌");
    });

    it("should show ❓ when tool_result is missing", () => {
      const logContent = [
        JSON.stringify({ type: "tool_use", tool_name: "get_commit", tool_id: "gc_001", parameters: { sha: "abc" } }),
        JSON.stringify({ type: "result", status: "success", stats: { total_tokens: 100, input_tokens: 80, output_tokens: 20, cached: 0, duration_ms: 1000, tool_calls: 1 } }),
      ].join("\n");

      const result = parseGeminiLog(logContent);

      expect(result.markdown).toContain("❓");
    });

    it("should render the Information section from result stats", () => {
      const logContent = [
        JSON.stringify({ type: "message", role: "assistant", content: "Done.", delta: true }),
        JSON.stringify({
          type: "result",
          status: "success",
          stats: { total_tokens: 321230, input_tokens: 315667, output_tokens: 911, cached: 245217, duration_ms: 72546, tool_calls: 7 },
        }),
      ].join("\n");

      const result = parseGeminiLog(logContent);

      expect(result.markdown).toContain("## 📊 Information");
      expect(result.markdown).toContain("321,230");
      expect(result.markdown).toContain("315,667");
      expect(result.markdown).toContain("911");
      expect(result.markdown).toContain("245,217");
      expect(result.markdown).toContain("Tool Calls:** 7");
    });

    it("should render duration from result stats", () => {
      const logContent = [JSON.stringify({ type: "result", status: "success", stats: { total_tokens: 100, input_tokens: 80, output_tokens: 20, cached: 0, duration_ms: 90000, tool_calls: 0 } })].join("\n");

      const result = parseGeminiLog(logContent);

      // 90000ms = 1m 30s
      expect(result.markdown).toContain("1m 30s");
    });

    it("should skip non-JSON lines like runner wrapper output", () => {
      const logContent = `[INFO] Starting containers...
[WARN] ⚠️  Using --env-all
[SUCCESS] Containers started
${JSON.stringify({ type: "message", role: "assistant", content: "I will proceed.", delta: true })}
[INFO] Executing agent command...
${JSON.stringify({ type: "result", status: "success", stats: { total_tokens: 50, input_tokens: 40, output_tokens: 10, cached: 0, duration_ms: 2000, tool_calls: 0 } })}`;

      const result = parseGeminiLog(logContent);

      expect(result.markdown).toContain("I will proceed.");
      expect(result.markdown).toContain("## 📊 Information");
    });

    it("should show 'No commands or tools used' when there are no tool calls", () => {
      const logContent = [
        JSON.stringify({ type: "message", role: "assistant", content: "Task complete.", delta: true }),
        JSON.stringify({ type: "result", status: "success", stats: { total_tokens: 50, input_tokens: 40, output_tokens: 10, cached: 0, duration_ms: 1000, tool_calls: 0 } }),
      ].join("\n");

      const result = parseGeminiLog(logContent);

      expect(result.markdown).toContain("No commands or tools used.");
    });

    it("should always return empty mcpFailures and maxTurnsHit=false", () => {
      const logContent = JSON.stringify({ type: "result", status: "success", stats: { total_tokens: 10, input_tokens: 8, output_tokens: 2, cached: 0, duration_ms: 500, tool_calls: 0 } });

      const result = parseGeminiLog(logContent);

      expect(result.mcpFailures).toEqual([]);
      expect(result.maxTurnsHit).toBe(false);
    });
  });

  describe("formatGeminiToolCall function", () => {
    it("should format a successful tool call with parameters and output", () => {
      const params = JSON.stringify({ owner: "github", repo: "gh-aw" }, null, 2);
      const output = '{"items": []}';
      const result = formatGeminiToolCall("list_pull_requests", params, output, "✅");

      expect(result).toContain("list_pull_requests");
      expect(result).toContain("✅");
      expect(result).toContain("Parameters");
      expect(result).toContain("Output");
    });

    it("should format a failed tool call", () => {
      const result = formatGeminiToolCall("create_issue", "{}", "Permission denied", "❌");

      expect(result).toContain("❌");
      expect(result).toContain("create_issue");
    });

    it("should handle empty parameters and output", () => {
      const result = formatGeminiToolCall("noop", "", "", "✅");

      expect(result).toContain("noop");
      expect(result).toContain("✅");
    });
  });
});
