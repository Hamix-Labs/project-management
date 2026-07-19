import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import type { GitBranch, GitWorktree, GitWorktreeCheckoutStatus } from "@/types/git";
import { worktreeGitCopy } from "../worktreeGitCopy";
import { WorktreeRow } from "./WorktreeRow";

const branch: GitBranch = {
  id: "00000000-0000-4000-8000-000000000040",
  repository_id: "00000000-0000-4000-8000-000000000010",
  name: "hamix/task-abcd1234",
  head_sha: "abc123",
  created_at: "2026-06-22T12:00:00Z",
};

const worktree: GitWorktree = {
  id: "00000000-0000-4000-8000-000000000030",
  repository_id: "00000000-0000-4000-8000-000000000010",
  path: "/repo/.hamix/worktrees/hamix-task-abcd1234",
  name: "hamix-task-abcd1234",
  is_main: false,
  branch_id: branch.id,
  created_at: "2026-06-22T12:00:00Z",
};

const primaryWorktree: GitWorktree = {
  ...worktree,
  id: "00000000-0000-4000-8000-000000000031",
  path: "/repo/main",
  name: "Hamix",
  is_main: true,
  branch_id: "00000000-0000-4000-8000-000000000041",
};

const cleanCheckoutStatus: GitWorktreeCheckoutStatus = {
  worktree_id: worktree.id,
  available: true,
  dirty: false,
  detached: false,
  head_commit_at: "2026-06-22T12:00:00Z",
};

describe("WorktreeRow", () => {
  it("renders managed worktree summary with branch pill", () => {
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

    expect(screen.getByText("hamix-task-abcd1234")).toBeInTheDocument();
    expect(screen.queryByText(worktreeGitCopy.primaryWorktreeBadge)).not.toBeInTheDocument();
    expect(screen.getByText("hamix/task-abcd1234")).toBeInTheDocument();
    expect(screen.getByText(worktreeGitCopy.statusClean)).toBeInTheDocument();
    expect(
      screen.getByRole("button", { name: /Worktree actions for hamix-task-abcd1234/i }),
    ).toBeInTheDocument();
  });

  it("offers unregister and delete for managed worktrees", async () => {
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

    await user.click(
      screen.getByRole("button", { name: /Worktree actions for hamix-task-abcd1234/i }),
    );
    expect(screen.getByRole("menuitem", { name: /Unregister worktree/i })).toBeInTheDocument();
    expect(screen.getByRole("menuitem", { name: /Delete worktree/i })).toBeInTheDocument();
  });

  it("hides the actions menu for the primary checkout", () => {
    render(
      <ul>
        <WorktreeRow
          worktree={primaryWorktree}
          branches={[]}
          onUnregister={vi.fn()}
          onDeleteFromDisk={vi.fn()}
        />
      </ul>,
    );

    expect(screen.getByText(worktreeGitCopy.primaryWorktreeBadge)).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: /Worktree actions/i })).not.toBeInTheDocument();
    expect(screen.queryByRole("menuitem", { name: /Unregister/i })).not.toBeInTheDocument();
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

    await user.click(screen.getByRole("button", { name: /expand hamix-task-abcd1234/i }));

    expect(screen.getByText(worktreeGitCopy.locationLabel)).toBeInTheDocument();
    expect(screen.getByText("/repo/.hamix/worktrees/hamix-task-abcd1234")).toBeInTheDocument();
    expect(screen.getAllByText("hamix/task-abcd1234").length).toBeGreaterThanOrEqual(2);

    await user.click(screen.getByRole("button", { name: /collapse hamix-task-abcd1234/i }));
    expect(screen.queryByText(worktreeGitCopy.locationLabel)).not.toBeInTheDocument();
  });
});
