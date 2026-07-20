import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import { ApiError } from "@/api";
import { DeleteConfirmDialog } from "./DeleteConfirmDialog";

describe("DeleteConfirmDialog", () => {
  it("shows repository delete confirmation copy", () => {
    render(
      <DeleteConfirmDialog
        target={{
          kind: "repository",
          id: "repo-1",
          label: "/repo/main",
          repositoryId: "repo-1",
        }}
        pending={false}
        error={null}
        onClose={() => {}}
        onConfirm={() => {}}
      />,
    );

    expect(screen.getByRole("heading", { name: /delete repository/i })).toBeInTheDocument();
    expect(screen.getByText(/repo\/main/i)).toBeInTheDocument();
    expect(screen.getByText(/files on disk are not deleted/i)).toBeInTheDocument();
  });

  it("disables confirm when a running task blocks delete", async () => {
    const user = userEvent.setup();
    const onConfirm = vi.fn();
    const blocked = new ApiError("A task is still running", {
      status: 409,
      code: "has_running_task",
    });

    render(
      <DeleteConfirmDialog
        target={{
          kind: "repository",
          id: "repo-1",
          label: "/repo/main",
          repositoryId: "repo-1",
        }}
        pending={false}
        error={blocked}
        onClose={() => {}}
        onConfirm={onConfirm}
      />,
    );

    const confirm = screen.getByRole("button", { name: /^delete$/i });
    expect(confirm).toBeDisabled();
    await user.click(confirm);
    expect(onConfirm).not.toHaveBeenCalled();
  });
});
