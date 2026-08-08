import { renderHook, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import {
  resolvePromptEditorFileWorktree,
  usePromptEditorFileWorktree,
} from "./usePromptEditorFileWorktree";

const { listGlobalGitWorktrees } = vi.hoisted(() => ({
  listGlobalGitWorktrees: vi.fn(),
}));

vi.mock("@/api/gitGlobal", () => ({
  listGlobalGitWorktrees,
}));

describe("resolvePromptEditorFileWorktree", () => {
  it("prefers an explicit task worktree over repository worktrees", () => {
    expect(
      resolvePromptEditorFileWorktree(" task-wt ", [
        { id: "main-wt", is_main: true },
      ]),
    ).toBe("task-wt");
  });

  it("falls back to the repository main worktree", () => {
    expect(
      resolvePromptEditorFileWorktree(undefined, [
        { id: "feature-wt", is_main: false },
        { id: "main-wt", is_main: true },
      ]),
    ).toBe("main-wt");
  });
});

describe("usePromptEditorFileWorktree", () => {
  beforeEach(() => {
    listGlobalGitWorktrees.mockReset();
  });

  it("does not fetch repository worktrees when a task worktree exists", () => {
    const { result } = renderHook(() =>
      usePromptEditorFileWorktree({
        worktreeId: " task-wt ",
        repositoryId: "repo-1",
      }),
    );

    expect(result.current).toEqual({
      worktreeId: "task-wt",
      resolving: false,
    });
    expect(listGlobalGitWorktrees).not.toHaveBeenCalled();
  });

  it("resolves the repository main worktree when the task worktree is missing", async () => {
    listGlobalGitWorktrees.mockResolvedValue([
      { id: "feature-wt", is_main: false },
      { id: "main-wt", is_main: true },
    ]);

    const { result } = renderHook(() =>
      usePromptEditorFileWorktree({ repositoryId: " repo-1 " }),
    );

    expect(result.current.resolving).toBe(true);

    await waitFor(() =>
      expect(result.current).toEqual({
        worktreeId: "main-wt",
        resolving: false,
      }),
    );
    expect(listGlobalGitWorktrees).toHaveBeenCalledWith(
      "repo-1",
      expect.objectContaining({ signal: expect.any(AbortSignal) }),
    );
  });
});
