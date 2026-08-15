import { screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { expect } from "vitest";

export async function waitForCreateTaskEnabled(dialog: HTMLElement) {
  await waitFor(() => {
    expect(
      within(dialog).getByRole("button", { name: /^create task$/i }),
    ).not.toBeDisabled();
  });
}

/** Opens create compose from home. Includes the portaled sticky action bar. */
export async function openNewTaskModal(
  user: ReturnType<typeof userEvent.setup>,
) {
  await user.click(screen.getByRole("button", { name: /\+?\s*new task/i }));
  await screen.findByRole(
    "heading",
    { name: /^new task$/i },
    { timeout: 10_000 },
  );
  await screen.findByTestId("task-compose-sticky-footer");
  return document.body;
}

export async function choosePriorityInDialog(
  user: ReturnType<typeof userEvent.setup>,
  dialog: HTMLElement,
  level: "low" | "medium" | "high" | "critical" = "medium",
) {
  const labels = {
    low: /^low$/i,
    medium: /^medium$/i,
    high: /^high$/i,
    critical: /^critical$/i,
  } as const;
  const group = within(dialog).getByRole("radiogroup", {
    name: /^priority$/i,
  });
  await user.click(within(group).getByRole("radio", { name: labels[level] }));
}

export async function addCriterionInDialog(
  user: ReturnType<typeof userEvent.setup>,
  dialog: HTMLElement,
  text: string,
) {
  const addBtn =
    within(dialog).queryByTestId("compose-criteria-add") ??
    within(dialog).getByRole("button", { name: /new criterion/i });
  await user.click(addBtn);
  const criterionDialog = await screen.findByRole("dialog", {
    name: /new criterion/i,
  });
  await user.type(
    within(criterionDialog).getByLabelText(/^criterion$/i),
    text,
  );
  await user.click(
    within(criterionDialog).getByRole("button", { name: /^add criterion$/i }),
  );
}
