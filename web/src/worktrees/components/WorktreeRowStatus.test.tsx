import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import type { GitWorktreeCheckoutStatus } from "@/types/git";
import { worktreeGitCopy } from "../worktreeGitCopy";
import { WorktreeRowStatus } from "./WorktreeRowStatus";

vi.mock("@/shared/useNow", () => ({
  useNow: () => new Date("2026-07-04T18:00:00.000Z").getTime(),
}));

const availableClean: GitWorktreeCheckoutStatus = {
  worktree_id: "00000000-0000-4000-8000-000000000030",
  available: true,
  dirty: false,
  detached: false,
  head_commit_at: "2026-07-04T16:00:00.000Z",
};

describe("WorktreeRowStatus", () => {
  it("shows clean label with prose relative time", () => {
    render(<WorktreeRowStatus checkoutStatus={availableClean} />);
    expect(screen.getByText(worktreeGitCopy.statusClean)).toBeInTheDocument();
    expect(screen.getByText(worktreeGitCopy.statusLastCommit("2 hours ago"))).toBeInTheDocument();
  });

  it("shows uncommitted changes when dirty", () => {
    render(
      <WorktreeRowStatus
        checkoutStatus={{
          ...availableClean,
          dirty: true,
          head_commit_at: "2026-07-04T17:48:00.000Z",
        }}
      />,
    );
    expect(screen.getByText(worktreeGitCopy.statusDirty)).toBeInTheDocument();
    expect(screen.getByText(worktreeGitCopy.statusLastCommit("12 minutes ago"))).toBeInTheDocument();
  });

  it("shows unavailable dash with title", () => {
    render(
      <WorktreeRowStatus
        checkoutStatus={{
          worktree_id: availableClean.worktree_id,
          available: false,
          reason: "path_missing",
        }}
      />,
    );
    expect(screen.getByText(worktreeGitCopy.statusUnavailable)).toBeInTheDocument();
    expect(screen.getByTitle(worktreeGitCopy.statusUnavailableTitle)).toBeInTheDocument();
  });
});
