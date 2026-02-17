import { describe, it, expect, beforeEach, afterEach, vi } from "vitest";
import fs from "fs";
import path from "path";

describe("parse_openclaw_log.cjs", () => {
  let mockCore, originalConsole, originalProcess;
  let main, parseOpenClawLog, formatOpenClawToolCall;

  beforeEach(async () => {
    originalConsole = global.console;
    originalProcess = { ...process };
    global.console = { log: vi.fn(), error: vi.fn() };

    mockCore = {
      debug: vi.fn(),
      info: vi.fn(),
      notice: vi.fn(),
      warning: vi.fn(),
      error: vi.fn(),
      setFailed: vi.fn(),
      setOutput: vi.fn(),
      exportVariable: vi.fn(),
      setSecret: vi.fn(),
      getInput: vi.fn(),
      getBooleanInput: vi.fn(),
      getMultilineInput: vi.fn(),
      getState: vi.fn(),
      saveState: vi.fn(),
      startGroup: vi.fn(),
      endGroup: vi.fn(),
      group: vi.fn(),
      addPath: vi.fn(),
      setCommandEcho: vi.fn(),
      isDebug: vi.fn().mockReturnValue(false),
      getIDToken: vi.fn(),
      toPlatformPath: vi.fn(),
      toPosixPath: vi.fn(),
      toWin32Path: vi.fn(),
      summary: { addRaw: vi.fn().mockReturnThis(), write: vi.fn().mockResolvedValue() },
    };

    global.core = mockCore;

    // Import the module to get the exported functions
    const module = await import("./parse_openclaw_log.cjs?" + Date.now());
    main = module.main;
    parseOpenClawLog = module.parseOpenClawLog;
    formatOpenClawToolCall = module.formatOpenClawToolCall;
  });

  afterEach(() => {
    delete process.env.GH_AW_AGENT_OUTPUT;
    global.console = originalConsole;
    process.env = originalProcess.env;
    delete global.core;
  });

  const runScript = async logContent => {
    const tempFile = path.join(process.cwd(), `test_openclaw_log_${Date.now()}.txt`);
    fs.writeFileSync(tempFile, logContent);
    process.env.GH_AW_AGENT_OUTPUT = tempFile;
    try {
      await main();
    } finally {
      if (fs.existsSync(tempFile)) {
        fs.unlinkSync(tempFile);
      }
    }
  };

  describe("parseOpenClawLog function", () => {
    it("should handle empty log content", () => {
      const result = parseOpenClawLog("");
      expect(result.markdown).toContain("No log content provided");
      expect(result.logEntries).toEqual([]);
      expect(result.mcpFailures).toEqual([]);
      expect(result.maxTurnsHit).toBe(false);
    });

    it("should handle null log content", () => {
      const result = parseOpenClawLog(null);
      expect(result.markdown).toContain("No log content provided");
      expect(result.logEntries).toEqual([]);
    });

    it("should parse tool_call JSON entries", () => {
      const logContent = '{"type":"tool_call","name":"bash","input":{"command":"echo hello"}}\n{"type":"tool_result","result":"hello"}';
      const result = parseOpenClawLog(logContent);

      expect(result.markdown).toContain("bash");
      expect(result.markdown).toContain("Tool Calls");
      expect(result.logEntries.length).toBeGreaterThan(0);
    });

    it("should parse tool_use JSON entries", () => {
      const logContent = '{"type":"tool_use","name":"github_search","input":{"query":"test"}}';
      const result = parseOpenClawLog(logContent);

      expect(result.markdown).toContain("github_search");
      expect(result.markdown).toContain("Tool Calls:** 1");
    });

    it("should parse message/reasoning entries", () => {
      const logContent = '{"type":"message","content":"I will analyze the repository structure to understand the codebase."}';
      const result = parseOpenClawLog(logContent);

      expect(result.markdown).toContain("analyze the repository");
      expect(result.logEntries.length).toBe(1);
      expect(result.logEntries[0].type).toBe("assistant");
    });

    it("should parse thinking entries", () => {
      const logContent = '{"type":"thinking","text":"Let me consider the approach carefully before proceeding."}';
      const result = parseOpenClawLog(logContent);

      expect(result.markdown).toContain("consider the approach");
    });

    it("should parse reasoning entries", () => {
      const logContent = '{"type":"reasoning","message":"The best approach is to use the existing API."}';
      const result = parseOpenClawLog(logContent);

      expect(result.markdown).toContain("best approach");
    });

    it("should skip short messages (10 chars or less)", () => {
      const logContent = '{"type":"message","content":"ok"}';
      const result = parseOpenClawLog(logContent);

      // Short messages are skipped
      expect(result.logEntries.length).toBe(0);
    });

    it("should handle error entries", () => {
      const logContent = '{"type":"error","message":"Connection timed out after 30s"}';
      const result = parseOpenClawLog(logContent);

      expect(result.markdown).toContain("Error");
      expect(result.markdown).toContain("Connection timed out");
    });

    it("should handle error entries with error field", () => {
      const logContent = '{"type":"error","error":"Invalid API key"}';
      const result = parseOpenClawLog(logContent);

      expect(result.markdown).toContain("Invalid API key");
    });

    it("should track MCP server connections", () => {
      const logContent = '{"type":"mcp_connected","server":"github","name":"github"}\n{"type":"mcp_init","name":"safeoutputs"}';
      const result = parseOpenClawLog(logContent);

      expect(result.markdown).toContain("MCP Server connected");
      expect(result.markdown).toContain("github");
      expect(result.mcpFailures).toEqual([]);
    });

    it("should track MCP server failures", () => {
      const logContent = '{"type":"mcp_failed","name":"broken_server","error":"Connection refused"}\n{"type":"mcp_error","server":"another_broken","error":"Timeout"}';
      const result = parseOpenClawLog(logContent);

      expect(result.markdown).toContain("MCP Server failed");
      expect(result.markdown).toContain("broken_server");
      expect(result.mcpFailures).toContain("broken_server");
      expect(result.mcpFailures).toContain("another_broken");
    });

    it("should track token usage", () => {
      const logContent = '{"type":"usage","total_tokens":1500}\n{"type":"token_usage","tokens":2000}';
      const result = parseOpenClawLog(logContent);

      expect(result.markdown).toContain("Total Tokens Used");
      expect(result.markdown).toContain("2,000");
    });

    it("should use highest token count", () => {
      const logContent = '{"type":"usage","total_tokens":500}\n{"type":"usage","total_tokens":1000}\n{"type":"usage","total_tokens":800}';
      const result = parseOpenClawLog(logContent);

      expect(result.markdown).toContain("1,000");
    });

    it("should handle tool_result entries", () => {
      const logContent = '{"type":"tool_call","name":"bash","input":{"command":"ls"}}\n{"type":"tool_result","result":"file1.txt\\nfile2.txt"}';
      const result = parseOpenClawLog(logContent);

      expect(result.logEntries.length).toBe(2);
      expect(result.logEntries[1].type).toBe("user");
      expect(result.logEntries[1].message.content[0].type).toBe("tool_result");
    });

    it("should handle tool_result with error flag", () => {
      const logContent = '{"type":"tool_result","result":"Permission denied","is_error":true}';
      const result = parseOpenClawLog(logContent);

      const toolResultEntry = result.logEntries.find(e => e.type === "user");
      expect(toolResultEntry.message.content[0].is_error).toBe(true);
    });

    it("should handle tool_result with output field", () => {
      const logContent = '{"type":"tool_result","output":"some output data"}';
      const result = parseOpenClawLog(logContent);

      expect(result.logEntries.length).toBe(1);
      expect(result.logEntries[0].message.content[0].content).toBe("some output data");
    });

    it("should handle tool_result with object result", () => {
      const logContent = '{"type":"tool_result","result":{"status":"ok","data":[1,2,3]}}';
      const result = parseOpenClawLog(logContent);

      expect(result.logEntries.length).toBe(1);
      const content = result.logEntries[0].message.content[0].content;
      expect(content).toContain("status");
    });

    it("should skip malformed JSON lines that start with { or [", () => {
      const logContent = '{broken json\n[also broken\n{"type":"message","content":"This is a valid longer message entry"}';
      const result = parseOpenClawLog(logContent);

      // Only the valid JSON should be parsed
      expect(result.logEntries.length).toBe(1);
    });

    it("should include substantive non-JSON text in reasoning", () => {
      const logContent = 'This is a substantive text output from the agent that should be included\n{"type":"message","content":"And here is a JSON message with more content"}';
      const result = parseOpenClawLog(logContent);

      expect(result.markdown).toContain("substantive text output");
    });

    it("should skip timestamp-like non-JSON lines", () => {
      const logContent = '2024-01-15T10:30:00Z some log message\n{"type":"message","content":"This is the actual content from the agent"}';
      const result = parseOpenClawLog(logContent);

      expect(result.markdown).not.toContain("2024-01-15");
    });

    it("should skip short non-JSON lines (20 chars or less)", () => {
      const logContent = 'debug info\n{"type":"message","content":"Longer actual content from the agent for testing"}';
      const result = parseOpenClawLog(logContent);

      expect(result.markdown).not.toContain("debug info");
    });

    it("should handle multiple tool calls and count them", () => {
      const logContent = ['{"type":"tool_call","name":"bash","input":{"command":"ls"}}', '{"type":"tool_call","name":"github_search","input":{"query":"test"}}', '{"type":"tool_call","name":"bash","input":{"command":"cat file.txt"}}'].join(
        "\n"
      );
      const result = parseOpenClawLog(logContent);

      expect(result.markdown).toContain("Tool Calls:** 3");
    });

    it("should show no tool calls message when none are present", () => {
      const logContent = '{"type":"message","content":"I analyzed the situation and came to a conclusion without tools"}';
      const result = parseOpenClawLog(logContent);

      expect(result.markdown).toContain("No tool calls detected");
    });

    it("should handle tool_call with function field for name", () => {
      const logContent = '{"type":"tool_call","function":"my_function","arguments":{"arg1":"value1"}}';
      const result = parseOpenClawLog(logContent);

      expect(result.markdown).toContain("my_function");
    });

    it("should handle tool_call with tool field for name", () => {
      const logContent = '{"type":"tool_call","tool":"my_tool","params":{"key":"value"}}';
      const result = parseOpenClawLog(logContent);

      expect(result.markdown).toContain("my_tool");
    });

    it("should handle tool_call with string params", () => {
      const logContent = '{"type":"tool_call","name":"eval","input":"print(42)"}';
      const result = parseOpenClawLog(logContent);

      expect(result.logEntries.length).toBe(1);
      const toolUse = result.logEntries[0].message.content[0];
      expect(toolUse.input).toEqual({ params: "print(42)" });
    });

    it("should handle unknown entry types gracefully", () => {
      const logContent = '{"type":"unknown_type","data":"some data"}\n{"type":"message","content":"This is valid content from the agent"}';
      const result = parseOpenClawLog(logContent);

      // Should not crash, just skip unknown types
      expect(result.logEntries.length).toBe(1);
    });

    it("should handle entry with event field instead of type", () => {
      const logContent = '{"event":"tool_call","name":"test_tool","input":{"x":1}}';
      const result = parseOpenClawLog(logContent);

      expect(result.markdown).toContain("test_tool");
    });

    it("should always return maxTurnsHit as false", () => {
      const logContent = '{"type":"message","content":"OpenClaw uses timeout-based execution, not turn-based limiting"}';
      const result = parseOpenClawLog(logContent);

      // OpenClaw uses timeout, not max-turns
      expect(result.maxTurnsHit).toBe(false);
    });
  });

  describe("formatOpenClawToolCall function", () => {
    it("should format a basic tool call", () => {
      const result = formatOpenClawToolCall("bash", '{"command":"ls -la"}', "", "⏳");

      expect(result).toContain("bash");
      expect(result).toContain("Parameters");
    });

    it("should include response when provided", () => {
      const result = formatOpenClawToolCall("test_tool", '{"arg":"value"}', '{"result":"ok"}', "✅");

      expect(result).toContain("test_tool");
      expect(result).toContain("Parameters");
      expect(result).toContain("Response");
    });

    it("should include token estimate", () => {
      const result = formatOpenClawToolCall("tool", '{"data":"test"}', "", "⏳");

      expect(result).toContain("~");
      expect(result).toContain("t</code>");
    });

    it("should handle empty params", () => {
      const result = formatOpenClawToolCall("tool", "", "", "✅");

      // No parameters section when params is empty
      expect(result).not.toContain("Parameters");
    });
  });

  describe("main function integration", () => {
    it("should handle valid log file with tool calls", async () => {
      const validLog = [
        '{"type":"message","content":"I will analyze the repository and find relevant files for you."}',
        '{"type":"tool_call","name":"bash","input":{"command":"ls -la"}}',
        '{"type":"tool_result","result":"file1.txt\\nfile2.txt"}',
        '{"type":"usage","total_tokens":500}',
      ].join("\n");

      await runScript(validLog);

      expect(mockCore.summary.addRaw).toHaveBeenCalled();
      expect(mockCore.summary.write).toHaveBeenCalled();
      expect(mockCore.setFailed).not.toHaveBeenCalled();
    });

    it("should handle log with MCP failures", async () => {
      const logWithFailures = ['{"type":"mcp_failed","name":"broken_server","error":"Connection refused"}', '{"type":"message","content":"I encountered an error connecting to the MCP server."}'].join("\n");

      await runScript(logWithFailures);

      expect(mockCore.summary.addRaw).toHaveBeenCalled();
      expect(mockCore.summary.write).toHaveBeenCalled();
      expect(mockCore.setFailed).toHaveBeenCalledWith("MCP server(s) failed to launch: broken_server");
    });

    it("should handle multiple MCP failures", async () => {
      const logWithMultipleFailures = ['{"type":"mcp_failed","name":"server1","error":"Timeout"}', '{"type":"mcp_error","name":"server2","error":"Auth failed"}'].join("\n");

      await runScript(logWithMultipleFailures);

      expect(mockCore.setFailed).toHaveBeenCalledWith("MCP server(s) failed to launch: server1, server2");
    });

    it("should handle missing log file", async () => {
      process.env.GH_AW_AGENT_OUTPUT = "/nonexistent/file.log";
      await main();
      expect(mockCore.info).toHaveBeenCalledWith("Log path not found: /nonexistent/file.log");
      expect(mockCore.setFailed).not.toHaveBeenCalled();
    });

    it("should handle missing environment variable", async () => {
      delete process.env.GH_AW_AGENT_OUTPUT;
      await main();
      expect(mockCore.info).toHaveBeenCalledWith("No agent log file specified");
      expect(mockCore.setFailed).not.toHaveBeenCalled();
    });

    it("should handle log with only non-JSON content", async () => {
      await runScript("plain text output from the agent");

      // OpenClaw parser should handle plain text gracefully (unlike Claude which requires structured entries)
      expect(mockCore.summary.addRaw).toHaveBeenCalled();
      expect(mockCore.summary.write).toHaveBeenCalled();
    });

    it("should handle mixed JSON and non-JSON content", async () => {
      const mixedLog = [
        "Starting OpenClaw agent execution...",
        '{"type":"message","content":"I will help you with the task at hand by analyzing the code."}',
        "Processing request...",
        '{"type":"tool_call","name":"bash","input":{"command":"echo test"}}',
        '{"type":"tool_result","result":"test"}',
        "Completed successfully",
      ].join("\n");

      await runScript(mixedLog);

      expect(mockCore.summary.addRaw).toHaveBeenCalled();
      expect(mockCore.summary.write).toHaveBeenCalled();
      expect(mockCore.setFailed).not.toHaveBeenCalled();
    });

    it("should handle empty log file", async () => {
      await runScript("");

      expect(mockCore.summary.addRaw).toHaveBeenCalled();
      expect(mockCore.summary.write).toHaveBeenCalled();
    });
  });
});
