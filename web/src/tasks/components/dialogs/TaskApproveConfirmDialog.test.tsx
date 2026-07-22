import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import { TaskApproveConfirmDialog } from "./TaskApproveConfirmDialog";

describe("TaskApproveConfirmDialog", () => {
  it("confirms approve", async () => {
    const user = userEvent.setup();
    const onConfirm = vi.fn();
    render(
      <TaskApproveConfirmDialog
        taskTitle="Ship gate"
        saving={false}
        pending={false}
        onCancel={vi.fn()}
        onConfirm={onConfirm}
      />,
    );
    expect(screen.getByRole("heading", { name: /approve this task/i })).toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: /^approve$/i }));
    expect(onConfirm).toHaveBeenCalledOnce();
  });
});
