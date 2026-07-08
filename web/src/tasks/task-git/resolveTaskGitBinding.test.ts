import { beforeEach, describe, expect, it, vi } from "vitest";
import {
  FACTORY_GIT_BRANCH_ID,
  FACTORY_GIT_REPO_ID,
  FACTORY_GIT_WORKTREE_ID,
  gitBranchFactory,
  gitRepositoryFactory,
  gitWorktreeFactory,
} from "@/test/factories/git";
import { resolveTaskGitBinding } from "./resolveTaskGitBinding";

const { mockListWorktrees, mockListBranches } = vi.hoisted(() => ({
  mockListWorktrees: vi.fn(),
  mockListBranches: vi.fn(),
}));

vi.mock("@/api/gitGlobal", () => ({
  listGlobalGitWorktrees: mockListWorktrees,
  listGlobalGitBranches: mockListBranches,
}));

describe("resolveTaskGitBinding", () => {
  beforeEach(() => {
    mockListWorktrees.mockReset();
    mockListBranches.mockReset();
  });

  it("returns null when worktree id is empty", async () => {
    await expect(
      resolveTaskGitBinding("", [gitRepositoryFactory()]),
    ).resolves.toBeNull();
    expect(mockListWorktrees).not.toHaveBeenCalled();
  });

  it("resolves repo, worktree path, and branch name", async () => {
    mockListWorktrees.mockResolvedValue([
      gitWorktreeFactory({
        id: FACTORY_GIT_WORKTREE_ID,
        path: "/repo/feature",
        branch_id: FACTORY_GIT_BRANCH_ID,
      }),
    ]);
    mockListBranches.mockResolvedValue([
      gitBranchFactory({ id: FACTORY_GIT_BRANCH_ID, name: "feature/commits" }),
    ]);

    await expect(
      resolveTaskGitBinding(FACTORY_GIT_WORKTREE_ID, [
        gitRepositoryFactory({ id: FACTORY_GIT_REPO_ID, path: "/repo/main" }),
      ]),
    ).resolves.toEqual({
      repo: "/repo/main",
      worktree: "/repo/feature",
      branch: "feature/commits",
    });
  });

  it("searches the hinted repository first", async () => {
    const otherRepoId = "00000000-0000-4000-8000-000000000099";
    mockListWorktrees.mockImplementation(async (repositoryId: string) => {
      if (repositoryId === FACTORY_GIT_REPO_ID) {
        return [
          gitWorktreeFactory({
            id: FACTORY_GIT_WORKTREE_ID,
            path: "/repo/main",
          }),
        ];
      }
      return [];
    });
    mockListBranches.mockResolvedValue([
      gitBranchFactory({ name: "main" }),
    ]);

    await resolveTaskGitBinding(
      FACTORY_GIT_WORKTREE_ID,
      [
        gitRepositoryFactory({ id: otherRepoId, path: "/other" }),
        gitRepositoryFactory({ id: FACTORY_GIT_REPO_ID, path: "/repo/main" }),
      ],
      { repositoryIdHint: FACTORY_GIT_REPO_ID },
    );

    expect(mockListWorktrees.mock.calls[0]?.[0]).toBe(FACTORY_GIT_REPO_ID);
  });
});
