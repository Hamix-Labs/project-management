import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import { CloseConfirmDialog } from "./CloseConfirmDialog";

const BASE = {
  taskTitle: "My task",
  taskId: "abcdef12-3456-7890",
  taskNumber: 7 as number | null,
  saving: false,
  closePending: false,
};

describe("CloseConfirmDialog", () => {
  it("renders the reversible-close copy referencing #N", () => {
    render(
      <CloseConfirmDialog
        {...BASE}
        onCancel={vi.fn()}
        onConfirm={vi.fn()}
      />,
    );
    expect(
      screen.getByRole("dialog", {
        description: /stop execution and be marked closed/i,
      }),
    ).toBeInTheDocument();
    expect(
      screen.getByText(/stops execution and closes #7\. you can reopen later\./i),
    ).toBeInTheDocument();
  });

  it("falls back to the shortened UUID when no #N is available", () => {
    render(
      <CloseConfirmDialog
        {...BASE}
        taskNumber={null}
        onCancel={vi.fn()}
        onConfirm={vi.fn()}
      />,
    );
    expect(
      screen.getByText(/stops execution and closes abcdef12/i),
    ).toBeInTheDocument();
  });

  it("calls onCancel when Cancel is clicked", async () => {
    const user = userEvent.setup();
    const onCancel = vi.fn();
    render(
      <CloseConfirmDialog
        {...BASE}
        onCancel={onCancel}
        onConfirm={vi.fn()}
      />,
    );
    await user.click(screen.getByRole("button", { name: /^cancel$/i }));
    expect(onCancel).toHaveBeenCalledTimes(1);
  });

  it("calls onCancel on Escape (dismissible while busy)", async () => {
    const user = userEvent.setup();
    const onCancel = vi.fn();
    render(
      <CloseConfirmDialog
        {...BASE}
        closePending
        onCancel={onCancel}
        onConfirm={vi.fn()}
      />,
    );
    await user.keyboard("{Escape}");
    expect(onCancel).toHaveBeenCalledTimes(1);
  });

  it("does not render an alert region when error is null", () => {
    render(
      <CloseConfirmDialog
        {...BASE}
        onCancel={vi.fn()}
        onConfirm={vi.fn()}
      />,
    );
    expect(screen.queryByRole("alert")).not.toBeInTheDocument();
  });

  it("renders the underlying close error message when error is set", () => {
    render(
      <CloseConfirmDialog
        {...BASE}
        error="task busy: cannot close"
        onCancel={vi.fn()}
        onConfirm={vi.fn()}
      />,
    );
    expect(screen.getByRole("alert")).toHaveTextContent(
      /task busy: cannot close/i,
    );
  });

  it("keeps action buttons enabled when an error is showing so the user can retry", () => {
    render(
      <CloseConfirmDialog
        {...BASE}
        error="boom"
        onCancel={vi.fn()}
        onConfirm={vi.fn()}
      />,
    );
    expect(
      screen.getByRole("button", { name: /^cancel$/i }),
    ).not.toBeDisabled();
    expect(
      screen.getByRole("button", { name: /^close task$/i }),
    ).not.toBeDisabled();
  });

  it("renders the busy spinner overlay while pending", () => {
    render(
      <CloseConfirmDialog
        {...BASE}
        saving
        closePending
        onCancel={vi.fn()}
        onConfirm={vi.fn()}
      />,
    );
    expect(screen.getByRole("status")).toBeInTheDocument();
  });
});
