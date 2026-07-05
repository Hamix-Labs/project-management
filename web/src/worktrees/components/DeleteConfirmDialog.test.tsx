import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import { ApiError } from "@/api";
import { worktreeGitCopy } from "../worktreeGitCopy";
import { DeleteConfirmDialog } from "./DeleteConfirmDialog";

describe("DeleteConfirmDialog", () => {
  it("shows unregister copy for inventory-only mode", () => {
    render(
      <DeleteConfirmDialog
        target={{
          kind: "worktree",
          mode: "unregister",
          id: "wt-1",
          label: "feature-a",
          repositoryId: "repo-1",
        }}
        pending={false}
        error={null}
        onClose={() => {}}
        onConfirm={() => {}}
      />,
    );

    expect(
      screen.getByRole("heading", { name: worktreeGitCopy.unregisterWorktreeConfirmTitle }),
    ).toBeInTheDocument();
    expect(screen.getByText(/removed from Hamix only/i)).toBeInTheDocument();
    expect(screen.getByText(worktreeGitCopy.unregisterWorktreeConfirmFootnote)).toBeInTheDocument();
  });

  it("shows delete copy for remove_from_disk mode", () => {
    render(
      <DeleteConfirmDialog
        target={{
          kind: "worktree",
          mode: "remove_from_disk",
          id: "wt-1",
          label: "feature-a",
          repositoryId: "repo-1",
        }}
        pending={false}
        error={null}
        onClose={() => {}}
        onConfirm={() => {}}
      />,
    );

    expect(
      screen.getByRole("heading", { name: worktreeGitCopy.deleteWorktreeConfirmTitle }),
    ).toBeInTheDocument();
    expect(screen.getByText(/deleted from disk/i)).toBeInTheDocument();
    expect(screen.getByText(worktreeGitCopy.deleteWorktreeConfirmFootnote)).toBeInTheDocument();
  });

  it("offers force remove when delete fails on dirty worktree", async () => {
    const user = userEvent.setup();
    const onConfirm = vi.fn();
    const dirtyError = new ApiError("worktree has uncommitted changes; use force", {
      status: 409,
      code: "path_exists",
    });

    render(
      <DeleteConfirmDialog
        target={{
          kind: "worktree",
          mode: "remove_from_disk",
          id: "wt-1",
          label: "feature-a",
          repositoryId: "repo-1",
        }}
        pending={false}
        error={dirtyError}
        onClose={() => {}}
        onConfirm={onConfirm}
      />,
    );

    const confirm = screen.getByRole("button", { name: /delete worktree/i });
    expect(confirm).toBeDisabled();
    await user.click(screen.getByRole("checkbox", { name: /force remove/i }));
    expect(confirm).toBeEnabled();
    await user.click(confirm);
    expect(onConfirm).toHaveBeenCalledWith({ force: true });
  });
});
