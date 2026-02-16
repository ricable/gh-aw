import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import fs from "fs";
import path from "path";
import os from "os";

const core = {
  info: vi.fn(),
  warning: vi.fn(),
  error: vi.fn(),
  setFailed: vi.fn(),
};
global.core = core;

const { validateAndNormalizePath, validatePathWithinBase, validateDirectory, safeJoin } = require("./path_helpers.cjs");

describe("path_helpers", () => {
  let tempDir;

  beforeEach(() => {
    tempDir = fs.mkdtempSync(path.join(os.tmpdir(), "path-helpers-test-"));
    vi.clearAllMocks();
  });

  afterEach(() => {
    if (tempDir && fs.existsSync(tempDir)) {
      fs.rmSync(tempDir, { recursive: true, force: true });
    }
  });

  describe("validateAndNormalizePath", () => {
    it("should normalize absolute paths", () => {
      const testPath = "/home/user/file.txt";
      const result = validateAndNormalizePath(testPath, "test file");
      expect(result).toBe(path.normalize(path.resolve(testPath)));
    });

    it("should normalize relative paths to absolute", () => {
      const testPath = "./file.txt";
      const result = validateAndNormalizePath(testPath, "test file");
      expect(path.isAbsolute(result)).toBe(true);
    });

    it("should trim whitespace from paths", () => {
      const testPath = "  /home/user/file.txt  ";
      const result = validateAndNormalizePath(testPath, "test file");
      expect(result).toBe(path.normalize(path.resolve("/home/user/file.txt")));
    });

    it("should throw on null bytes", () => {
      expect(() => validateAndNormalizePath("/home/user/file\0.txt", "test file")).toThrow(
        "Security: test file contains null bytes"
      );
    });

    it("should throw on empty string", () => {
      expect(() => validateAndNormalizePath("", "test file")).toThrow(
        "Invalid test file: path must be a non-empty string"
      );
    });

    it("should throw on whitespace-only string", () => {
      expect(() => validateAndNormalizePath("   ", "test file")).toThrow(
        "Invalid test file: path cannot be empty or whitespace-only"
      );
    });

    it("should throw on null input", () => {
      expect(() => validateAndNormalizePath(null, "test file")).toThrow(
        "Invalid test file: path must be a non-empty string"
      );
    });

    it("should throw on undefined input", () => {
      expect(() => validateAndNormalizePath(undefined, "test file")).toThrow(
        "Invalid test file: path must be a non-empty string"
      );
    });

    it("should normalize paths with ..", () => {
      const testPath = "/home/user/../admin/file.txt";
      const result = validateAndNormalizePath(testPath, "test file");
      expect(result).toBe(path.normalize(path.resolve(testPath)));
    });

    it("should normalize paths with multiple slashes", () => {
      const testPath = "/home//user///file.txt";
      const result = validateAndNormalizePath(testPath, "test file");
      expect(result).toBe(path.normalize(path.resolve("/home/user/file.txt")));
    });
  });

  describe("validatePathWithinBase", () => {
    it("should allow paths within base directory", () => {
      const baseDir = tempDir;
      const filePath = path.join(tempDir, "file.txt");
      const result = validatePathWithinBase(filePath, baseDir, "test file");
      expect(result).toBe(path.normalize(filePath));
    });

    it("should allow relative paths within base directory", () => {
      const baseDir = tempDir;
      // Create a subdirectory to test relative paths
      const subdir = path.join(tempDir, "subdir");
      fs.mkdirSync(subdir);
      
      // Change to base directory for relative path testing
      const originalCwd = process.cwd();
      process.chdir(baseDir);
      
      try {
        const result = validatePathWithinBase("subdir/file.txt", baseDir, "test file");
        expect(result).toBe(path.normalize(path.join(baseDir, "subdir/file.txt")));
      } finally {
        process.chdir(originalCwd);
      }
    });

    it("should reject paths that escape base directory with ../", () => {
      const baseDir = path.join(tempDir, "base");
      fs.mkdirSync(baseDir);
      const escapePath = path.join(baseDir, "../outside.txt");
      
      expect(() => validatePathWithinBase(escapePath, baseDir, "test file")).toThrow(
        /Security: test file must be within/
      );
    });

    it("should reject paths that escape base directory with ../../", () => {
      const baseDir = path.join(tempDir, "base");
      fs.mkdirSync(baseDir);
      const escapePath = path.join(baseDir, "../../etc/passwd");
      
      expect(() => validatePathWithinBase(escapePath, baseDir, "test file")).toThrow(
        /Security: test file must be within/
      );
    });

    it("should reject absolute paths outside base directory", () => {
      const baseDir = path.join(tempDir, "base");
      fs.mkdirSync(baseDir);
      const outsidePath = "/etc/passwd";
      
      expect(() => validatePathWithinBase(outsidePath, baseDir, "test file")).toThrow(
        /Security: test file must be within/
      );
    });

    it("should allow nested paths within base directory", () => {
      const baseDir = tempDir;
      const nestedPath = path.join(tempDir, "a/b/c/file.txt");
      const result = validatePathWithinBase(nestedPath, baseDir, "test file");
      expect(result).toBe(path.normalize(nestedPath));
    });

    it("should handle paths with . segments", () => {
      const baseDir = tempDir;
      const pathWithDots = path.join(tempDir, "./subdir/./file.txt");
      const result = validatePathWithinBase(pathWithDots, baseDir, "test file");
      expect(result).toBe(path.normalize(path.join(tempDir, "subdir/file.txt")));
    });
  });

  describe("validateDirectory", () => {
    it("should validate existing directory", () => {
      const result = validateDirectory(tempDir, "test directory");
      expect(result).toBe(path.normalize(path.resolve(tempDir)));
    });

    it("should throw on non-existent directory when createIfMissing is false", () => {
      const nonExistent = path.join(tempDir, "nonexistent");
      expect(() => validateDirectory(nonExistent, "test directory", false)).toThrow(
        "test directory does not exist"
      );
    });

    it("should create directory when createIfMissing is true", () => {
      const newDir = path.join(tempDir, "newdir");
      const result = validateDirectory(newDir, "test directory", true);
      expect(fs.existsSync(newDir)).toBe(true);
      expect(fs.statSync(newDir).isDirectory()).toBe(true);
      expect(result).toBe(path.normalize(path.resolve(newDir)));
    });

    it("should throw when path is a file not a directory", () => {
      const filePath = path.join(tempDir, "file.txt");
      fs.writeFileSync(filePath, "content");
      
      expect(() => validateDirectory(filePath, "test directory")).toThrow(
        "test directory is not a directory"
      );
    });

    it("should create nested directories when createIfMissing is true", () => {
      const nestedDir = path.join(tempDir, "a/b/c");
      const result = validateDirectory(nestedDir, "test directory", true);
      expect(fs.existsSync(nestedDir)).toBe(true);
      expect(fs.statSync(nestedDir).isDirectory()).toBe(true);
      expect(result).toBe(path.normalize(path.resolve(nestedDir)));
    });
  });

  describe("safeJoin", () => {
    it("should join and normalize path segments", () => {
      const result = safeJoin("/home", "user", "file.txt");
      expect(result).toBe(path.normalize("/home/user/file.txt"));
    });

    it("should normalize redundant separators", () => {
      const result = safeJoin("/home//user", "///file.txt");
      expect(result).toBe(path.normalize("/home/user/file.txt"));
    });

    it("should handle . segments", () => {
      const result = safeJoin("/home", "./user", "./file.txt");
      expect(result).toBe(path.normalize("/home/user/file.txt"));
    });

    it("should handle .. segments", () => {
      const result = safeJoin("/home/user", "../admin", "file.txt");
      expect(result).toBe(path.normalize("/home/admin/file.txt"));
    });

    it("should work with single segment", () => {
      const result = safeJoin("/home");
      expect(result).toBe(path.normalize("/home"));
    });

    it("should work with relative paths", () => {
      const result = safeJoin("home", "user", "file.txt");
      expect(result).toBe(path.normalize("home/user/file.txt"));
    });
  });

  describe("security scenarios", () => {
    it("should prevent null byte injection in file paths", () => {
      const maliciousPath = `/home/user/file.txt\0/../../etc/passwd`;
      expect(() => validateAndNormalizePath(maliciousPath, "malicious file")).toThrow(
        "Security: malicious file contains null bytes"
      );
    });

    it("should prevent directory traversal with mixed .. and valid segments", () => {
      const baseDir = path.join(tempDir, "base");
      fs.mkdirSync(baseDir);
      const traversalPath = path.join(baseDir, "subdir/../../outside.txt");
      
      expect(() => validatePathWithinBase(traversalPath, baseDir, "traversal file")).toThrow(
        /Security: traversal file must be within/
      );
    });

    it("should handle Windows-style paths with backslashes", () => {
      const baseDir = tempDir;
      const winPath = path.join(tempDir, "subdir\\file.txt");
      const result = validatePathWithinBase(winPath, baseDir, "windows file");
      // On Unix, backslashes are valid filename characters, not separators
      // On Windows, they are path separators and get normalized
      expect(result).toBe(path.normalize(path.resolve(winPath)));
    });
  });
});
