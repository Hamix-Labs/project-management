import { describe, expect, it, vi, beforeEach } from "vitest";
import { resolveMentionWorktreeId } from "./resolveMentionWorktreeId";

vi.mock("@/api/gitGlobal", () => ({
  listGlobalGitWorktrees: vi.fn(),
}));

import { listGlobalGitWorktrees } from "@/api/gitGlobal";

const listMock = vi.mocked(listGlobalGitWorktrees);

describe("resolveMentionWorktreeId", () => {
  beforeEach(() => {
    listMock.mockReset();
  });

  it("prefers an explicit worktree id", async () => {
    await expect(
      resolveMentionWorktreeId({
        worktreeId: " wt-1 ",
        repositoryId: "repo-1",
      }),
    ).resolves.toBe("wt-1");
    expect(listMock).not.toHaveBeenCalled();
  });

  it("falls back to the repository main worktree", async () => {
    listMock.mockResolvedValue([
      {
        id: "wt-side",
        repository_id: "repo-1",
        path: "/tmp/side",
        host_path: "",
        name: "side",
        is_main: false,
        created_at: "",
      },
      {
        id: "wt-main",
        repository_id: "repo-1",
        path: "/tmp/main",
        host_path: "",
        name: "main",
        is_main: true,
        created_at: "",
      },
    ]);
    await expect(
      resolveMentionWorktreeId({ repositoryId: "repo-1" }),
    ).resolves.toBe("wt-main");
  });

  it("returns null when neither worktree nor repository is set", async () => {
    await expect(resolveMentionWorktreeId({})).resolves.toBeNull();
  });

  it("returns null when listing fails", async () => {
    listMock.mockRejectedValue(new Error("network"));
    await expect(
      resolveMentionWorktreeId({ repositoryId: "repo-1" }),
    ).resolves.toBeNull();
  });
});
