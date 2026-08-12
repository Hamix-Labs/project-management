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

/** Opens create compose page from home and returns the page root. */
export async function openNewTaskModal(
  user: ReturnType<typeof userEvent.setup>,
) {
  await user.click(screen.getByRole("button", { name: /\+?\s*new task/i }));
  const heading = await screen.findByRole(
    "heading",
    { name: /^new task$/i },
    { timeout: 10_000 },
  );
  const page = heading.closest(".task-compose-page");
  if (!(page instanceof HTMLElement)) {
    throw new Error("compose page root not found");
  }
  return page;
}

export async function choosePriorityInDialog(
  user: ReturnType<typeof userEvent.setup>,
  dialog: HTMLElement,
  level: "low" | "medium" | "high" | "critical" = "medium",
) {
  const combo = within(dialog).getByRole("combobox", {
    name: /^priority$/i,
  });
  await user.click(combo);
  await user.click(
    screen.getByRole("option", { name: new RegExp(`^${level}$`, "i") }),
  );
}

export async function addCriterionInDialog(
  user: ReturnType<typeof userEvent.setup>,
  dialog: HTMLElement,
  text: string,
) {
  await user.click(
    within(dialog).getByRole("button", { name: /new criterion/i }),
  );
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
