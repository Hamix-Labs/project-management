import { render, screen, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router-dom";
import { describe, expect, it, vi } from "vitest";
import { TaskPolishDialog } from "./TaskPolishDialog";

vi.mock("@/tasks/prompt-editor/useOpenPromptEditor", () => ({
  useOpenPolishPromptEditor: () =>
    vi.fn(() => {
      /* navigation mocked */
    }),
}));

function renderPolish(
  props: Partial<React.ComponentProps<typeof TaskPolishDialog>> = {},
) {
  return render(
    <MemoryRouter>
      <TaskPolishDialog
        taskId="task-1"
        saving={false}
        pending={false}
        onCancel={vi.fn()}
        onConfirm={vi.fn()}
        {...props}
      />
    </MemoryRouter>,
  );
}

describe("TaskPolishDialog", () => {
  it("requires non-empty instructions before confirm", async () => {
    const user = userEvent.setup();
    const onConfirm = vi.fn();
    renderPolish({
      onConfirm,
      initialInstructions: "<p>tighten spacing</p>",
    });
    expect(screen.getByRole("button", { name: /^polish$/i })).toBeEnabled();
    await user.click(screen.getByRole("button", { name: /^polish$/i }));
    expect(onConfirm).toHaveBeenCalledWith({
      instructions: "<p>tighten spacing</p>",
      flaggedCriterionIds: [],
      newCriteria: [],
    });
  });

  it("keeps submit disabled when only flags or drafts are present", async () => {
    const user = userEvent.setup();
    const onConfirm = vi.fn();
    renderPolish({
      onConfirm,
      criteria: [{ id: "c1", text: "Auth works" }],
    });
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

  it("exposes Open Prompt Editor instead of an in-place editor", () => {
    renderPolish();
    expect(
      screen.getByRole("button", { name: /open prompt editor/i }),
    ).toBeInTheDocument();
    expect(screen.queryByRole("textbox")).toBeNull();
  });
});
