import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import { TaskDetailToolbarActions } from "./TaskDetailAttentionBar";

describe("TaskDetailToolbarActions", () => {
  it("invokes edit and close handlers", async () => {
    const user = userEvent.setup();
    const onEdit = vi.fn();
    const onClose = vi.fn();
    render(
      <TaskDetailToolbarActions
        saving={false}
        onEdit={onEdit}
        onClose={onClose}
      />,
    );

    await user.click(screen.getByRole("button", { name: /edit task/i }));
    await user.click(screen.getByRole("button", { name: /^close$/i }));
    expect(onEdit).toHaveBeenCalledOnce();
    expect(onClose).toHaveBeenCalledOnce();
  });

  it("invokes enqueue when worktree is ready", async () => {
    const user = userEvent.setup();
    const onEnqueue = vi.fn();
    render(
      <TaskDetailToolbarActions
        saving={false}
        onEdit={vi.fn()}
        onEnqueue={onEnqueue}
      />,
    );
    await user.click(screen.getByRole("button", { name: /enqueue task/i }));
    expect(onEnqueue).toHaveBeenCalledOnce();
  });

  it("disables enqueue while worktree is provisioning", () => {
    render(
      <TaskDetailToolbarActions
        saving={false}
        onEdit={vi.fn()}
        onEnqueue={vi.fn()}
        enqueueDisabledReason="Worktree is still provisioning"
      />,
    );
    expect(screen.getByRole("button", { name: /enqueue task/i })).toBeDisabled();
  });

  it("disables action buttons while saving", () => {
    render(
      <TaskDetailToolbarActions
        saving
        onEdit={vi.fn()}
        onClose={vi.fn()}
      />,
    );

    expect(screen.getByRole("button", { name: /edit task/i })).toBeDisabled();
    expect(screen.getByRole("button", { name: /^close$/i })).toBeDisabled();
  });

  it("disables edit but keeps close enabled when canEdit is false", async () => {
    const user = userEvent.setup();
    const onEdit = vi.fn();
    const onClose = vi.fn();
    render(
      <TaskDetailToolbarActions
        saving={false}
        canEdit={false}
        onEdit={onEdit}
        onClose={onClose}
      />,
    );

    const editButton = screen.getByRole("button", { name: /edit task/i });
    expect(editButton).toBeDisabled();
    await user.click(editButton);
    expect(onEdit).not.toHaveBeenCalled();

    await user.click(screen.getByRole("button", { name: /^close$/i }));
    expect(onClose).toHaveBeenCalledOnce();
  });

  it("swaps Close for Reopen when onReopen is provided", async () => {
    const user = userEvent.setup();
    const onReopen = vi.fn();
    render(
      <TaskDetailToolbarActions
        saving={false}
        onEdit={vi.fn()}
        onReopen={onReopen}
      />,
    );
    expect(
      screen.queryByRole("button", { name: /^close$/i }),
    ).not.toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: /^reopen$/i }));
    expect(onReopen).toHaveBeenCalledOnce();
  });

  it("shows Reopening… pending label", () => {
    render(
      <TaskDetailToolbarActions
        saving={false}
        onEdit={vi.fn()}
        onReopen={vi.fn()}
        reopenPending
      />,
    );
    expect(
      screen.getByRole("button", { name: /^reopening…$/i }),
    ).toBeDisabled();
  });

  it("renders Approve & Open PR and Polish when review handlers are provided", async () => {
    const user = userEvent.setup();
    const onOpenPr = vi.fn();
    const onPolish = vi.fn();
    render(
      <TaskDetailToolbarActions
        saving={false}
        onEdit={vi.fn()}
        onClose={vi.fn()}
        onOpenPr={onOpenPr}
        onPolish={onPolish}
      />,
    );

    await user.click(screen.getByRole("button", { name: /approve & open pr/i }));
    await user.click(screen.getByRole("button", { name: /^polish$/i }));
    expect(onOpenPr).toHaveBeenCalledOnce();
    expect(onPolish).toHaveBeenCalledOnce();
  });

  it("renders Mark done when approve handler is provided", async () => {
    const user = userEvent.setup();
    const onApprove = vi.fn();
    render(
      <TaskDetailToolbarActions
        saving={false}
        onEdit={vi.fn()}
        onClose={vi.fn()}
        onApprove={onApprove}
      />,
    );
    await user.click(screen.getByRole("button", { name: /^mark done$/i }));
    expect(onApprove).toHaveBeenCalledOnce();
  });

  it("renders Model configuration only when showModelConfig is true", async () => {
    const user = userEvent.setup();
    const onConfigureModel = vi.fn();

    const { rerender } = render(
      <TaskDetailToolbarActions
        saving={false}
        onEdit={vi.fn()}
        onClose={vi.fn()}
        onConfigureModel={onConfigureModel}
      />,
    );
    expect(
      screen.queryByRole("button", { name: /model configuration/i }),
    ).not.toBeInTheDocument();

    rerender(
      <TaskDetailToolbarActions
        saving={false}
        onEdit={vi.fn()}
        onClose={vi.fn()}
        onConfigureModel={onConfigureModel}
        showModelConfig
      />,
    );

    const button = screen.getByRole("button", {
      name: /model configuration/i,
    });
    await user.click(button);
    expect(onConfigureModel).toHaveBeenCalledOnce();
  });

  describe("autonomy toggle", () => {
    it("does not render the toggle when autonomyMode is hidden / unset", () => {
      render(
        <TaskDetailToolbarActions
          saving={false}
          onEdit={vi.fn()}
          onClose={vi.fn()}
        />,
      );
      expect(
        screen.queryByRole("button", { name: /^(resume|pause)$/i }),
      ).not.toBeInTheDocument();
    });

    it("renders 'Pause' when autonomyMode=ready", async () => {
      const user = userEvent.setup();
      const onToggleAutonomy = vi.fn();
      render(
        <TaskDetailToolbarActions
          saving={false}
          onEdit={vi.fn()}
          onClose={vi.fn()}
          autonomyMode="ready"
          onToggleAutonomy={onToggleAutonomy}
        />,
      );
      const button = screen.getByRole("button", { name: /^pause$/i });
      await user.click(button);
      expect(onToggleAutonomy).toHaveBeenCalledOnce();
    });

    it("renders 'Resume' when autonomyMode=on_hold", async () => {
      const user = userEvent.setup();
      const onToggleAutonomy = vi.fn();
      render(
        <TaskDetailToolbarActions
          saving={false}
          onEdit={vi.fn()}
          onClose={vi.fn()}
          autonomyMode="on_hold"
          onToggleAutonomy={onToggleAutonomy}
        />,
      );
      const button = screen.getByRole("button", { name: /^resume$/i });
      await user.click(button);
      expect(onToggleAutonomy).toHaveBeenCalledOnce();
    });

    it("shows the pending label and disables the toggle while pending", () => {
      render(
        <TaskDetailToolbarActions
          saving={false}
          onEdit={vi.fn()}
          onClose={vi.fn()}
          autonomyMode="ready"
          onToggleAutonomy={vi.fn()}
          autonomyPending
        />,
      );
      const button = screen.getByRole("button", { name: /^pausing…$/i });
      expect(button).toBeDisabled();
    });
  });

  it("no longer renders the legacy inline model-config panel", () => {
    render(
      <TaskDetailToolbarActions
        saving={false}
        onEdit={vi.fn()}
        onClose={vi.fn()}
        onConfigureModel={vi.fn()}
        showModelConfig
      />,
    );

    expect(
      screen.queryByRole("heading", { name: /model configuration/i, level: 3 }),
    ).not.toBeInTheDocument();
    expect(screen.queryByText(/global model/i)).not.toBeInTheDocument();
    expect(screen.queryByText(/per-task model/i)).not.toBeInTheDocument();
  });
});
