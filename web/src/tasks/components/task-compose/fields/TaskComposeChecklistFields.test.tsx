import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import { TaskComposeChecklistFields } from "./TaskComposeChecklistFields";

describe("TaskComposeChecklistFields", () => {
  it("shows a check marker, verify command badge, and edit actions on each row", () => {
    render(
      <TaskComposeChecklistFields
        checklistHeadingId="checklist-heading"
        checklistItems={[
          {
            text: "The full test suite still passes.",
            verify_commands: [{ command: "go test ./...", expected_outcome: "pass" }],
          },
        ]}
        disabled={false}
        onOpenNewCriterion={vi.fn()}
        onOpenEditCriterion={vi.fn()}
        onRemoveRow={vi.fn()}
      />,
    );

    expect(screen.getByText("The full test suite still passes.")).toBeInTheDocument();
    expect(document.querySelector(".compose-criteria__check")).not.toBeNull();
    expect(
      screen.getByLabelText(/1 verify command for the execute agent/i),
    ).toHaveTextContent("1 command");
    expect(screen.getByRole("button", { name: "Edit" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /remove/i })).toBeInTheDocument();

    const row = document.querySelector(".task-checklist-row");
    const primary = row?.querySelector(".task-checklist-row-primary");
    const trailing = row?.querySelector(".task-checklist-row-trailing");
    const actions = row?.querySelector(".task-checklist-row-actions");
    expect(primary?.querySelector(".task-checklist-verify-badge")).toBeNull();
    expect(trailing?.querySelector(".task-checklist-verify-badge")).not.toBeNull();
    expect(actions).not.toBeNull();
    expect(trailing && actions && trailing.contains(actions)).toBe(true);
  });

  it("omits the verify badge when a criterion has no commands", () => {
    render(
      <TaskComposeChecklistFields
        checklistHeadingId="checklist-heading"
        checklistItems={[{ text: "The chosen entry point is named with a justification for why it was picked." }]}
        disabled={false}
        onOpenNewCriterion={vi.fn()}
        onOpenEditCriterion={vi.fn()}
        onRemoveRow={vi.fn()}
      />,
    );

    expect(
      screen.queryByLabelText(/verify command for the execute agent/i),
    ).not.toBeInTheDocument();
    const row = document.querySelector(".task-checklist-row");
    const trailing = row?.querySelector(".task-checklist-row-trailing");
    const actions = row?.querySelector(".task-checklist-row-actions");
    expect(trailing?.querySelector(".task-checklist-verify-badge")).toBeNull();
    expect(screen.getByRole("button", { name: "Edit" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /remove/i })).toBeInTheDocument();
    expect(trailing && actions && trailing.contains(actions)).toBe(true);
  });

  it("opens edit when anywhere on the criterion row is clicked", async () => {
    const user = userEvent.setup();
    const onOpenEditCriterion = vi.fn();
    const item = {
      text: "The chosen entry point is named with a justification for why it was picked.",
      verify_commands: [{ command: "go test ./...", expected_outcome: "pass" }],
    };

    render(
      <TaskComposeChecklistFields
        checklistHeadingId="checklist-heading"
        checklistItems={[item]}
        disabled={false}
        onOpenNewCriterion={vi.fn()}
        onOpenEditCriterion={onOpenEditCriterion}
        onRemoveRow={vi.fn()}
      />,
    );

    await user.click(
      screen.getByText(/The chosen entry point is named with a justification/i),
    );
    expect(onOpenEditCriterion).toHaveBeenCalledWith(0, item);
  });
});
