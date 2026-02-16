// @ts-check

import { describe, it, expect, beforeEach, afterEach, vi } from "vitest";

// Import the core shim
import coreShim from "./core_shim.cjs";

describe("core_shim", () => {
  let consoleLogSpy;
  let consoleWarnSpy;
  let consoleErrorSpy;
  let originalDebug;

  beforeEach(() => {
    // Spy on console methods
    consoleLogSpy = vi.spyOn(console, "log").mockImplementation(() => {});
    consoleWarnSpy = vi.spyOn(console, "warn").mockImplementation(() => {});
    consoleErrorSpy = vi.spyOn(console, "error").mockImplementation(() => {});

    // Save original DEBUG env var
    originalDebug = process.env.DEBUG;
  });

  afterEach(() => {
    // Restore console methods
    consoleLogSpy.mockRestore();
    consoleWarnSpy.mockRestore();
    consoleErrorSpy.mockRestore();

    // Restore DEBUG env var
    if (originalDebug === undefined) {
      delete process.env.DEBUG;
    } else {
      process.env.DEBUG = originalDebug;
    }
  });

  describe("info", () => {
    it("should log informational messages to console.log", () => {
      coreShim.info("Test information");

      expect(consoleLogSpy).toHaveBeenCalledWith("[INFO] Test information");
      expect(consoleLogSpy).toHaveBeenCalledTimes(1);
    });

    it("should handle empty messages", () => {
      coreShim.info("");

      expect(consoleLogSpy).toHaveBeenCalledWith("[INFO] ");
    });

    it("should handle multiline messages", () => {
      coreShim.info("Line 1\nLine 2");

      expect(consoleLogSpy).toHaveBeenCalledWith("[INFO] Line 1\nLine 2");
    });
  });

  describe("warning", () => {
    it("should log warning messages to console.warn", () => {
      coreShim.warning("Test warning");

      expect(consoleWarnSpy).toHaveBeenCalledWith("[WARNING] Test warning");
      expect(consoleWarnSpy).toHaveBeenCalledTimes(1);
    });

    it("should handle empty warnings", () => {
      coreShim.warning("");

      expect(consoleWarnSpy).toHaveBeenCalledWith("[WARNING] ");
    });
  });

  describe("error", () => {
    it("should log error messages to console.error", () => {
      coreShim.error("Test error");

      expect(consoleErrorSpy).toHaveBeenCalledWith("[ERROR] Test error");
      expect(consoleErrorSpy).toHaveBeenCalledTimes(1);
    });

    it("should handle empty errors", () => {
      coreShim.error("");

      expect(consoleErrorSpy).toHaveBeenCalledWith("[ERROR] ");
    });
  });

  describe("debug", () => {
    it("should log debug messages when DEBUG env var is set", () => {
      process.env.DEBUG = "1";
      coreShim.debug("Test debug");

      expect(consoleLogSpy).toHaveBeenCalledWith("[DEBUG] Test debug");
      expect(consoleLogSpy).toHaveBeenCalledTimes(1);
    });

    it("should not log debug messages when DEBUG env var is not set", () => {
      delete process.env.DEBUG;
      coreShim.debug("Test debug");

      expect(consoleLogSpy).not.toHaveBeenCalled();
    });

    it("should log debug messages when DEBUG env var is any truthy value", () => {
      process.env.DEBUG = "true";
      coreShim.debug("Test debug");

      expect(consoleLogSpy).toHaveBeenCalledWith("[DEBUG] Test debug");
    });
  });

  describe("setFailed", () => {
    it("should log failure messages to console.error", () => {
      coreShim.setFailed("Test failure");

      expect(consoleErrorSpy).toHaveBeenCalledWith("[FAILED] Test failure");
      expect(consoleErrorSpy).toHaveBeenCalledTimes(1);
    });

    it("should handle empty failure messages", () => {
      coreShim.setFailed("");

      expect(consoleErrorSpy).toHaveBeenCalledWith("[FAILED] ");
    });

    it("should not exit the process", () => {
      const exitSpy = vi.spyOn(process, "exit").mockImplementation(() => {});

      coreShim.setFailed("Test failure");

      expect(exitSpy).not.toHaveBeenCalled();
      exitSpy.mockRestore();
    });
  });

  describe("interface compatibility", () => {
    it("should expose all expected methods", () => {
      expect(typeof coreShim.info).toBe("function");
      expect(typeof coreShim.warning).toBe("function");
      expect(typeof coreShim.error).toBe("function");
      expect(typeof coreShim.debug).toBe("function");
      expect(typeof coreShim.setFailed).toBe("function");
    });
  });
});
