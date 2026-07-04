import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import type { GitBranch, GitWorktree, GitWorktreeCheckoutStatus } from "@/types/git";
import { worktreeGitCopy } from "../worktreeGitCopy";
import { WorktreeRow } from "./WorktreeRow";

const branch: GitBranch = {
  id: "00000000-0000-4000-8000-000000000040",
  repository_id: "00000000-0000-4000-8000-000000000010",
  name: "main",
  head_sha: "abc123",
  created_at: "2026-06-22T12:00:00Z",
};

const worktree: GitWorktree = {
  id: "00000000-0000-4000-8000-000000000030",
  repository_id: "00000000-0000-4000-8000-000000000010",
  path: "/repo/main",
  name: "Hamix",
  is_main: true,
  branch_id: branch.id,
  created_at: "2026-06-22T12:00:00Z",
};

const cleanCheckoutStatus: GitWorktreeCheckoutStatus = {
  worktree_id: worktree.id,
  available: true,
  dirty: false,
  detached: false,
  head_commit_at: "2026-06-22T12:00:00Z",
};

describe("WorktreeRow", () => {
  it("renders summary with branch pill and primary badge", () => {
    render(
      <ul>
        <WorktreeRow
          worktree={worktree}
          branches={[branch]}
          checkoutStatus={cleanCheckoutStatus}
          onUnregister={vi.fn()}
          onDeleteFromDisk={vi.fn()}
        />
      </ul>,
    );

    expect(screen.getByText("Hamix")).toBeInTheDocument();
    expect(screen.getByText(worktreeGitCopy.primaryWorktreeBadge)).toBeInTheDocument();
    expect(screen.getByText("main")).toBeInTheDocument();
    expect(screen.getByText(worktreeGitCopy.statusClean)).toBeInTheDocument();
  });

  it("toggles expand panel with location and branch", async () => {
    const user = userEvent.setup();
    render(
      <ul>
        <WorktreeRow
          worktree={worktree}
          branches={[branch]}
          checkoutStatus={cleanCheckoutStatus}
          onUnregister={vi.fn()}
          onDeleteFromDisk={vi.fn()}
        />
      </ul>,
    );

    expect(screen.queryByText(worktreeGitCopy.locationLabel)).not.toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: /expand hamix/i }));

    expect(screen.getByText(worktreeGitCopy.locationLabel)).toBeInTheDocument();
    expect(screen.getByText("/repo/main")).toBeInTheDocument();
    expect(screen.getAllByText("main").length).toBeGreaterThanOrEqual(2);

    await user.click(screen.getByRole("button", { name: /collapse hamix/i }));
    expect(screen.queryByText(worktreeGitCopy.locationLabel)).not.toBeInTheDocument();
  });
});
