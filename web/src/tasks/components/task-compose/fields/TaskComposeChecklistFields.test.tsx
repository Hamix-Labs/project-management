import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import { TaskComposeChecklistFields } from "./TaskComposeChecklistFields";

describe("TaskComposeChecklistFields", () => {
  it("shows verify command badge below criterion text in the meta row", () => {
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
    expect(
      screen.getByLabelText(/1 verify command for the execute agent/i),
    ).toHaveTextContent("1 command");
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
