// @ts-check
import { describe, it, expect, beforeEach, vi } from "vitest";

// Mock core global
const mockCore = {
  info: vi.fn(),
  warning: vi.fn(),
  error: vi.fn(),
};

global.core = mockCore;

describe("expired_entity_handlers", () => {
  let createExpiredEntityProcessor;

  beforeEach(async () => {
    vi.clearAllMocks();
    const module = await import("./expired_entity_handlers.cjs");
    createExpiredEntityProcessor = module.createExpiredEntityProcessor;
  });

  describe("createExpiredEntityProcessor", () => {
    it("should process entity with comment and close", async () => {
      const mockAddComment = vi.fn().mockResolvedValue({});
      const mockCloseEntity = vi.fn().mockResolvedValue({});
      const mockBuildMessage = vi.fn().mockReturnValue("Test closing message");

      const processor = createExpiredEntityProcessor("test-workflow", "https://github.com/test/repo/actions/runs/123", "test-workflow-id", {
        entityType: "issue",
        addComment: mockAddComment,
        closeEntity: mockCloseEntity,
        buildClosingMessage: mockBuildMessage,
      });

      const entity = {
        number: 42,
        url: "https://github.com/test/repo/issues/42",
        title: "Test Issue",
        expirationDate: new Date("2020-01-20T09:20:00.000Z"),
      };

      const result = await processor(entity);

      expect(result).toEqual({
        status: "closed",
        record: {
          number: 42,
          url: "https://github.com/test/repo/issues/42",
          title: "Test Issue",
        },
      });

      expect(mockBuildMessage).toHaveBeenCalledWith(entity, "test-workflow", "https://github.com/test/repo/actions/runs/123", "test-workflow-id");
      expect(mockAddComment).toHaveBeenCalledWith(entity, "Test closing message");
      expect(mockCloseEntity).toHaveBeenCalledWith(entity);
      expect(mockCore.info).toHaveBeenCalledWith("  Adding closing comment to issue #42");
      expect(mockCore.info).toHaveBeenCalledWith("  ✓ Comment added successfully");
      expect(mockCore.info).toHaveBeenCalledWith("  Closing issue #42");
      expect(mockCore.info).toHaveBeenCalledWith("  ✓ Issue closed successfully");
    });

    it("should skip entity when preCheck returns shouldSkip=true without closing", async () => {
      const mockAddComment = vi.fn();
      const mockCloseEntity = vi.fn();
      const mockBuildMessage = vi.fn();
      const mockPreCheck = vi.fn().mockResolvedValue({
        shouldSkip: true,
        reason: "Already closed",
      });

      const processor = createExpiredEntityProcessor("test-workflow", "https://github.com/test/repo/actions/runs/123", "test-workflow-id", {
        entityType: "discussion",
        addComment: mockAddComment,
        closeEntity: mockCloseEntity,
        buildClosingMessage: mockBuildMessage,
        preCheck: mockPreCheck,
      });

      const entity = {
        number: 99,
        url: "https://github.com/test/repo/discussions/99",
        title: "Already Closed",
        id: "D_test123",
      };

      const result = await processor(entity);

      expect(result).toEqual({
        status: "skipped",
        record: {
          number: 99,
          url: "https://github.com/test/repo/discussions/99",
          title: "Already Closed",
        },
      });

      expect(mockPreCheck).toHaveBeenCalledWith(entity);
      expect(mockCore.warning).toHaveBeenCalledWith("  Already closed");
      expect(mockAddComment).not.toHaveBeenCalled();
      expect(mockCloseEntity).not.toHaveBeenCalled();
    });

    it("should skip entity and close when preCheck returns shouldSkip=true with shouldClose=true", async () => {
      const mockAddComment = vi.fn();
      const mockCloseEntity = vi.fn().mockResolvedValue({});
      const mockBuildMessage = vi.fn();
      const mockPreCheck = vi.fn().mockResolvedValue({
        shouldSkip: true,
        shouldClose: true,
        reason: "Has existing comment, closing without adding another",
      });

      const processor = createExpiredEntityProcessor("test-workflow", "https://github.com/test/repo/actions/runs/123", "test-workflow-id", {
        entityType: "discussion",
        addComment: mockAddComment,
        closeEntity: mockCloseEntity,
        buildClosingMessage: mockBuildMessage,
        preCheck: mockPreCheck,
      });

      const entity = {
        number: 88,
        url: "https://github.com/test/repo/discussions/88",
        title: "Has Comment",
        id: "D_test456",
      };

      const result = await processor(entity);

      expect(result).toEqual({
        status: "skipped",
        record: {
          number: 88,
          url: "https://github.com/test/repo/discussions/88",
          title: "Has Comment",
        },
      });

      expect(mockPreCheck).toHaveBeenCalledWith(entity);
      expect(mockCore.warning).toHaveBeenCalledWith("  Has existing comment, closing without adding another");
      expect(mockCore.info).toHaveBeenCalledWith("  Attempting to close discussion #88 without adding another comment");
      expect(mockCloseEntity).toHaveBeenCalledWith(entity);
      expect(mockCore.info).toHaveBeenCalledWith("  ✓ Discussion closed successfully");
      expect(mockAddComment).not.toHaveBeenCalled();
    });

    it("should process entity when preCheck returns shouldSkip=false", async () => {
      const mockAddComment = vi.fn().mockResolvedValue({});
      const mockCloseEntity = vi.fn().mockResolvedValue({});
      const mockBuildMessage = vi.fn().mockReturnValue("Closing message");
      const mockPreCheck = vi.fn().mockResolvedValue({
        shouldSkip: false,
      });

      const processor = createExpiredEntityProcessor("test-workflow", "https://github.com/test/repo/actions/runs/123", "test-workflow-id", {
        entityType: "discussion",
        addComment: mockAddComment,
        closeEntity: mockCloseEntity,
        buildClosingMessage: mockBuildMessage,
        preCheck: mockPreCheck,
      });

      const entity = {
        number: 77,
        url: "https://github.com/test/repo/discussions/77",
        title: "No Comment",
        id: "D_test789",
      };

      const result = await processor(entity);

      expect(result).toEqual({
        status: "closed",
        record: {
          number: 77,
          url: "https://github.com/test/repo/discussions/77",
          title: "No Comment",
        },
      });

      expect(mockPreCheck).toHaveBeenCalledWith(entity);
      expect(mockAddComment).toHaveBeenCalledWith(entity, "Closing message");
      expect(mockCloseEntity).toHaveBeenCalledWith(entity);
    });

    it("should work without preCheck function", async () => {
      const mockAddComment = vi.fn().mockResolvedValue({});
      const mockCloseEntity = vi.fn().mockResolvedValue({});
      const mockBuildMessage = vi.fn().mockReturnValue("Simple close");

      const processor = createExpiredEntityProcessor("test-workflow", "https://github.com/test/repo/actions/runs/123", "test-workflow-id", {
        entityType: "pull request",
        addComment: mockAddComment,
        closeEntity: mockCloseEntity,
        buildClosingMessage: mockBuildMessage,
        // No preCheck provided
      });

      const entity = {
        number: 55,
        url: "https://github.com/test/repo/pull/55",
        title: "Simple PR",
      };

      const result = await processor(entity);

      expect(result).toEqual({
        status: "closed",
        record: {
          number: 55,
          url: "https://github.com/test/repo/pull/55",
          title: "Simple PR",
        },
      });

      expect(mockAddComment).toHaveBeenCalledWith(entity, "Simple close");
      expect(mockCloseEntity).toHaveBeenCalledWith(entity);
    });

    it("should capitalize entity type in log messages", async () => {
      const mockAddComment = vi.fn().mockResolvedValue({});
      const mockCloseEntity = vi.fn().mockResolvedValue({});
      const mockBuildMessage = vi.fn().mockReturnValue("Test");

      const processor = createExpiredEntityProcessor("test-workflow", "https://github.com/test/repo/actions/runs/123", "test-workflow-id", {
        entityType: "pull request",
        addComment: mockAddComment,
        closeEntity: mockCloseEntity,
        buildClosingMessage: mockBuildMessage,
      });

      const entity = {
        number: 1,
        url: "https://github.com/test/repo/pull/1",
        title: "Test",
      };

      await processor(entity);

      expect(mockCore.info).toHaveBeenCalledWith("  Adding closing comment to pull request #1");
      expect(mockCore.info).toHaveBeenCalledWith("  ✓ Pull request closed successfully");
    });

    it("should handle all three entity types correctly", async () => {
      const entities = [
        { type: "issue", number: 1, url: "https://github.com/test/repo/issues/1", title: "Issue" },
        { type: "pull request", number: 2, url: "https://github.com/test/repo/pull/2", title: "PR" },
        { type: "discussion", number: 3, url: "https://github.com/test/repo/discussions/3", title: "Discussion" },
      ];

      for (const entity of entities) {
        vi.clearAllMocks();

        const mockAddComment = vi.fn().mockResolvedValue({});
        const mockCloseEntity = vi.fn().mockResolvedValue({});
        const mockBuildMessage = vi.fn().mockReturnValue("Close message");

        const processor = createExpiredEntityProcessor("test-workflow", "https://github.com/test/repo/actions/runs/123", "test-workflow-id", {
          entityType: entity.type,
          addComment: mockAddComment,
          closeEntity: mockCloseEntity,
          buildClosingMessage: mockBuildMessage,
        });

        const result = await processor(entity);

        expect(result.status).toBe("closed");
        expect(result.record.number).toBe(entity.number);
        expect(mockAddComment).toHaveBeenCalled();
        expect(mockCloseEntity).toHaveBeenCalled();
      }
    });
  });
});
