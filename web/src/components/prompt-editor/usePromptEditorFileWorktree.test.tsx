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

    expect(result.current).toMatchObject({
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
      expect(result.current).toMatchObject({
        worktreeId: "main-wt",
        resolving: false,
      }),
    );
    expect(listGlobalGitWorktrees).toHaveBeenCalledWith(
      "repo-1",
      expect.objectContaining({ signal: expect.any(AbortSignal) }),
    );
  });

  it("reports unbound when there is neither a worktree nor a repository", () => {
    const { result } = renderHook(() => usePromptEditorFileWorktree({}));

    expect(result.current.gap).toBe("unbound");
    expect(result.current.resolving).toBe(false);
  });

  it("distinguishes a repository without a main worktree from a failed lookup", async () => {
    listGlobalGitWorktrees.mockResolvedValue([
      { id: "feature-wt", is_main: false },
    ]);
    const noMain = renderHook(() =>
      usePromptEditorFileWorktree({ repositoryId: "repo-1" }),
    );
    await waitFor(() => expect(noMain.result.current.resolving).toBe(false));
    expect(noMain.result.current.gap).toBe("no-main-worktree");

    listGlobalGitWorktrees.mockRejectedValue(new Error("offline"));
    const failed = renderHook(() =>
      usePromptEditorFileWorktree({ repositoryId: "repo-2" }),
    );
    await waitFor(() => expect(failed.result.current.resolving).toBe(false));
    expect(failed.result.current.gap).toBe("lookup-failed");
  });

  it("lets a caller await the in-flight lookup instead of seeing no worktree", async () => {
    let release: (worktrees: unknown) => void = () => {};
    listGlobalGitWorktrees.mockReturnValue(
      new Promise((resolve) => {
        release = resolve;
      }),
    );

    const { result } = renderHook(() =>
      usePromptEditorFileWorktree({ repositoryId: "repo-1" }),
    );

    expect(result.current.worktreeId).toBeUndefined();
    const awaited = result.current.whenResolved();

    release([{ id: "main-wt", is_main: true }]);
    await expect(awaited).resolves.toBe("main-wt");
  });

  it("resolves immediately against an explicit worktree", async () => {
    const { result } = renderHook(() =>
      usePromptEditorFileWorktree({ worktreeId: "task-wt" }),
    );

    await expect(result.current.whenResolved()).resolves.toBe("task-wt");
    expect(listGlobalGitWorktrees).not.toHaveBeenCalled();
  });
});
