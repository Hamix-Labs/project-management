import { describe, expect, it } from "vitest";
import { render, screen } from "@testing-library/react";
import type { GitWorktreeCheckoutStatus } from "@/types/git";
import { worktreeGitCopy } from "../worktreeGitCopy";
import { BranchSyncIndicator } from "./BranchSyncIndicator";

const baseStatus: GitWorktreeCheckoutStatus = {
  worktree_id: "00000000-0000-4000-8000-000000000030",
  available: true,
  dirty: false,
  detached: false,
  has_upstream: true,
  ahead: 0,
  behind: 0,
  upstream: "origin/main",
};

describe("BranchSyncIndicator", () => {
  it("shows Up to date when upstream matches", () => {
    render(<BranchSyncIndicator checkoutStatus={baseStatus} />);
    expect(screen.getByText(worktreeGitCopy.syncUpToDate)).toBeInTheDocument();
  });

  it("shows ahead and behind counts with spaces", () => {
    render(
      <BranchSyncIndicator
        checkoutStatus={{ ...baseStatus, ahead: 3, behind: 1 }}
      />,
    );
    expect(screen.getByText("↑ 3 ↓ 1")).toBeInTheDocument();
  });

  it("shows only behind when ahead is zero", () => {
    render(
      <BranchSyncIndicator checkoutStatus={{ ...baseStatus, ahead: 0, behind: 4 }} />,
    );
    expect(screen.getByText("↓ 4")).toBeInTheDocument();
  });

  it("renders nothing without upstream", () => {
    const { container } = render(
      <BranchSyncIndicator checkoutStatus={{ ...baseStatus, has_upstream: false }} />,
    );
    expect(container).toBeEmptyDOMElement();
  });
});
