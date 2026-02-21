import { describe, it, expect } from "vitest";

const { parseBoolTemplatable } = require("./templatable.cjs");

describe("templatable.cjs", () => {
  describe("parseBoolTemplatable", () => {
    it("returns defaultValue (true) for undefined", () => {
      expect(parseBoolTemplatable(undefined)).toBe(true);
    });

    it("returns defaultValue (true) for null", () => {
      expect(parseBoolTemplatable(null)).toBe(true);
    });

    it("respects a custom defaultValue", () => {
      expect(parseBoolTemplatable(undefined, false)).toBe(false);
      expect(parseBoolTemplatable(null, false)).toBe(false);
    });

    it("handles boolean true", () => {
      expect(parseBoolTemplatable(true)).toBe(true);
    });

    it("handles boolean false", () => {
      expect(parseBoolTemplatable(false)).toBe(false);
    });

    it('handles string "true"', () => {
      expect(parseBoolTemplatable("true")).toBe(true);
    });

    it('handles string "false"', () => {
      expect(parseBoolTemplatable("false")).toBe(false);
    });

    it("treats a resolved expression value other than false as truthy", () => {
      // GitHub Actions expressions that resolve to something other than "false"
      // (e.g. "yes", "1", an empty object representation) should be truthy.
      expect(parseBoolTemplatable("yes")).toBe(true);
      expect(parseBoolTemplatable("1")).toBe(true);
    });

    it("treats the string false-equivalent as falsy", () => {
      expect(parseBoolTemplatable("false", true)).toBe(false);
    });
  });
});
