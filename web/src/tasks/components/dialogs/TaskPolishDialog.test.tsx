import { render, screen, within } from "@testing-library/react";
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
    expect(onConfirm).toHaveBeenCalledWith({
      instructions: "  tighten spacing  ",
      flaggedCriterionIds: [],
      newCriteria: [],
    });
  });

  it("keeps submit disabled when only flags or drafts are present", async () => {
    const user = userEvent.setup();
    const onConfirm = vi.fn();
    render(
      <TaskPolishDialog
        saving={false}
        pending={false}
        criteria={[{ id: "c1", text: "Auth works" }]}
        onCancel={vi.fn()}
        onConfirm={onConfirm}
      />,
    );
    await user.click(screen.getByLabelText(/auth works/i));
    expect(screen.getByRole("button", { name: /^polish$/i })).toBeDisabled();
    await user.click(screen.getByRole("button", { name: /^add criterion$/i }));
    const sheet = document.querySelector(
      ".task-checklist-criterion-modal-sheet",
    );
    expect(sheet).not.toBeNull();
    await user.type(
      within(sheet as HTMLElement).getByLabelText(/criterion/i),
      "Docs",
    );
    await user.click(
      within(sheet as HTMLElement).getByRole("button", {
        name: /^add criterion$/i,
      }),
    );
    expect(screen.getByRole("button", { name: /^polish$/i })).toBeDisabled();
    expect(onConfirm).not.toHaveBeenCalled();
  });

  it("opens the checklist criterion modal to draft new criteria", async () => {
    const user = userEvent.setup();
    const onConfirm = vi.fn();
    render(
      <TaskPolishDialog
        saving={false}
        pending={false}
        criteria={[
          { id: "c1", text: "Auth works" },
          { id: "c2", text: "Tests pass" },
        ]}
        onCancel={vi.fn()}
        onConfirm={onConfirm}
      />,
    );

    expect(
      screen.getByText("not", { selector: ".task-polish-dialog__not" }),
    ).toBeInTheDocument();
    await user.click(screen.getByLabelText(/auth works/i));
    await user.click(screen.getByRole("button", { name: /^add criterion$/i }));
    const sheet = document.querySelector(
      ".task-checklist-criterion-modal-sheet",
    );
    expect(sheet).not.toBeNull();
    expect(
      within(sheet as HTMLElement).getByText(/one clear, testable requirement/i),
    ).toBeInTheDocument();
    await user.type(
      within(sheet as HTMLElement).getByLabelText(/criterion/i),
      "Docs updated",
    );
    await user.click(
      within(sheet as HTMLElement).getByRole("button", {
        name: /^add criterion$/i,
      }),
    );
    expect(screen.getByText("Docs updated")).toBeInTheDocument();
    expect(
      document.querySelector(".task-checklist-criterion-modal-sheet"),
    ).toBeNull();
    await user.type(
      screen.getByLabelText(/instructions/i),
      "fix the auth flow",
    );
    await user.click(screen.getByRole("button", { name: /^polish$/i }));
    expect(onConfirm).toHaveBeenCalledWith({
      instructions: "fix the auth flow",
      flaggedCriterionIds: ["c1"],
      newCriteria: [{ text: "Docs updated", verify_commands: [] }],
    });
  });

  it("removes a drafted criterion with the chip remove control", async () => {
    const user = userEvent.setup();
    render(
      <TaskPolishDialog
        saving={false}
        pending={false}
        onCancel={vi.fn()}
        onConfirm={vi.fn()}
      />,
    );
    await user.click(screen.getByRole("button", { name: /^add criterion$/i }));
    const sheet = document.querySelector(
      ".task-checklist-criterion-modal-sheet",
    ) as HTMLElement;
    await user.type(within(sheet).getByLabelText(/criterion/i), "Docs");
    await user.click(
      within(sheet).getByRole("button", { name: /^add criterion$/i }),
    );
    expect(screen.getByText("Docs")).toBeInTheDocument();
    await user.click(
      screen.getByRole("button", { name: /remove criterion docs/i }),
    );
    expect(screen.queryByText("Docs")).not.toBeInTheDocument();
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
    expect(
      screen.getByPlaceholderText(/describe what should change in this polish pass/i),
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
