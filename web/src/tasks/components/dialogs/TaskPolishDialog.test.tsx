import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import { TaskPolishDialog } from "./TaskPolishDialog";

vi.mock("@/components/rich-prompt", () => ({
  RichPromptEditor: ({
    id,
    value,
    onChange,
    disabled,
    placeholder,
  }: {
    id: string;
    value: string;
    onChange: (v: string) => void;
    disabled?: boolean;
    placeholder?: string;
  }) => (
    <textarea
      id={id}
      aria-labelledby={`${id}-label`}
      value={value}
      onChange={(e) => onChange(e.target.value)}
      disabled={disabled}
      placeholder={placeholder}
    />
  ),
}));

vi.mock("@/hooks/useProjectContextPromptBinding", () => ({
  useProjectContextPromptBinding: () => null,
}));

describe("TaskPolishDialog", () => {
  it("requires non-empty instructions before confirm", async () => {
    const user = userEvent.setup();
    const onConfirm = vi.fn();
    render(
      <TaskPolishDialog
        saving={false}
        pending={false}
        onCancel={vi.fn()}
        onConfirm={onConfirm}
      />,
    );
    expect(screen.getByRole("button", { name: /^polish$/i })).toBeDisabled();
    await user.type(
      screen.getByLabelText(/instructions/i),
      "  tighten spacing  ",
    );
    await user.click(screen.getByRole("button", { name: /^polish$/i }));
    expect(onConfirm).toHaveBeenCalledWith("  tighten spacing  ");
  });

  it("uses a wide modal shell for the rich instructions editor", () => {
    render(
      <TaskPolishDialog
        saving={false}
        pending={false}
        onCancel={vi.fn()}
        onConfirm={vi.fn()}
      />,
    );
    expect(screen.getByRole("dialog")).toHaveClass("modal-shell--wide");
  });

  it("renders inspiration chrome: title, close, @ hint, and Esc affordance", async () => {
    const user = userEvent.setup();
    const onCancel = vi.fn();
    render(
      <TaskPolishDialog
        saving={false}
        pending={false}
        onCancel={onCancel}
        onConfirm={vi.fn()}
      />,
    );

    expect(
      screen.getByRole("heading", { name: /^polish this task$/i }),
    ).toBeInTheDocument();
    expect(
      screen.getByText(/resume the existing agent conversation/i),
    ).toBeInTheDocument();
    expect(screen.getByText(/type/i).closest(".task-polish-dialog__hint")).toHaveTextContent(
      "Type @ to reference files",
    );
    expect(screen.getByText(/esc/i).closest(".task-polish-dialog__esc-hint")).toHaveTextContent(
      "Esc to cancel",
    );

    await user.click(screen.getByRole("button", { name: /^close$/i }));
    expect(onCancel).toHaveBeenCalledOnce();
  });
});
