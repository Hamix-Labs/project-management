import { beforeEach, describe, expect, it, vi } from "vitest";
import { FACTORY_GIT_WORKTREE_ID } from "@/test/factories/git";
import { resolveTaskGitBinding } from "./resolveTaskGitBinding";

const { mockGetWorktree } = vi.hoisted(() => ({
  mockGetWorktree: vi.fn(),
}));

vi.mock("@/api/gitGlobal", () => ({
  getGlobalGitWorktree: mockGetWorktree,
}));

describe("resolveTaskGitBinding", () => {
  beforeEach(() => {
    mockGetWorktree.mockReset();
  });

  it("returns null when worktree id is empty", async () => {
    await expect(resolveTaskGitBinding("")).resolves.toBeNull();
    expect(mockGetWorktree).not.toHaveBeenCalled();
  });

  it("resolves repo, worktree path, and branch name", async () => {
    mockGetWorktree.mockResolvedValue({
      id: FACTORY_GIT_WORKTREE_ID,
      repository_id: "repo",
      path: "/repo/feature",
      host_path: "",
      name: "feature",
      is_main: false,
      created_at: "2026-06-22T12:00:00Z",
      repository_path: "/repo/main",
      repository_host_path: "",
      branch_name: "feature/commits",
    });

    await expect(resolveTaskGitBinding(FACTORY_GIT_WORKTREE_ID)).resolves.toEqual({
      repo: "/repo/main",
      worktree: "/repo/feature",
      openPath: "/repo/feature",
      branch: "feature/commits",
    });
    expect(mockGetWorktree).toHaveBeenCalledWith(FACTORY_GIT_WORKTREE_ID, {
      signal: undefined,
    });
  });

  it("prefers host_path for openPath when present", async () => {
    mockGetWorktree.mockResolvedValue({
      id: FACTORY_GIT_WORKTREE_ID,
      repository_id: "repo",
      path: "/container/wt",
      host_path: "/Users/a/.hamix/wt",
      name: "feature",
      is_main: false,
      created_at: "2026-06-22T12:00:00Z",
      repository_path: "/repo/main",
      repository_host_path: "",
      branch_name: "feature/x",
    });

    await expect(resolveTaskGitBinding(FACTORY_GIT_WORKTREE_ID)).resolves.toMatchObject({
      worktree: "/container/wt",
      openPath: "/Users/a/.hamix/wt",
    });
  });

  it("returns null when get-by-id fails", async () => {
    mockGetWorktree.mockRejectedValue(new Error("not found"));
    await expect(resolveTaskGitBinding(FACTORY_GIT_WORKTREE_ID)).resolves.toBeNull();
  });
});
