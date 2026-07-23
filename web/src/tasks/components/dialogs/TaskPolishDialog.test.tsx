import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import { TaskPolishDialog } from "./TaskPolishDialog";

describe("TaskPolishDialog", () => {
  it("requires non-empty instructions before confirm", async () => {
    const user = userEvent.setup();
    const onConfirm = vi.fn();
    render(
      <TaskPolishDialog
        taskTitle="Ship gate"
        saving={false}
        pending={false}
        onCancel={vi.fn()}
        onConfirm={onConfirm}
      />,
    );
    expect(screen.getByRole("button", { name: /^polish$/i })).toBeDisabled();
    expect(screen.getByLabelText(/instructions/i)).toBeRequired();
    await user.type(
      screen.getByLabelText(/instructions/i),
      "  tighten spacing  ",
    );
    await user.click(screen.getByRole("button", { name: /^polish$/i }));
    expect(onConfirm).toHaveBeenCalledWith("tighten spacing");
  });
});
