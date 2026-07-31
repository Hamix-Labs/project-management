import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import type { TaskChecklistItemView } from "@/types";
import { TaskDetailChecklistItemList } from "./TaskDetailChecklistItemList";

const PENDING: TaskChecklistItemView = {
  id: "pending-1",
  sort_order: 1,
  text: "Hello World is written inside the 123.md file",
  done: false,
};

const DONE: TaskChecklistItemView = {
  id: "done-1",
  sort_order: 2,
  text: "123.md file created",
  done: true,
};

const DONE_VERIFIED: TaskChecklistItemView = {
  id: "done-verified-1",
  sort_order: 3,
  text: "The summary names the primary language, runtime, and main entry point.",
  done: true,
  verified_by: "execute_agent",
  evidence: "CODEBASE_TOUR.md names Go 1.25+ as the primary backend language.",
};

function renderList(
  items: TaskChecklistItemView[],
  overrides?: Partial<{
    taskStatus: import("@/types").Status;
    criteriaLocked: boolean;
    editCriterionPending: boolean;
    removeItemPending: boolean;
    addCriterionPending: boolean;
  }>,
) {
  const onOpenEditCriterionModal = vi.fn();
  const onRemoveChecklistItem = vi.fn();
  render(
    <TaskDetailChecklistItemList
      items={items}
      taskStatus={overrides?.taskStatus ?? "ready"}
      criteriaLocked={overrides?.criteriaLocked ?? false}
      editCriterionPending={overrides?.editCriterionPending ?? false}
      removeItemPending={overrides?.removeItemPending ?? false}
      addCriterionPending={overrides?.addCriterionPending ?? false}
      onOpenEditCriterionModal={onOpenEditCriterionModal}
      onRemoveChecklistItem={onRemoveChecklistItem}
    />,
  );
  return { onOpenEditCriterionModal, onRemoveChecklistItem };
}

describe("TaskDetailChecklistItemList", () => {
  // Pins the rule from the user-reported bug: once the agent has
  // marked a criterion done, the Edit affordance must lock — editing
  // the definition text after acceptance silently rewrites the
  // already-emitted checklist_item_toggled audit row's meaning. The
  // backend rejects this with ErrInvalidInput; this test guards the
  // UI half so users can't even attempt the bogus write.
  it("locks the Edit button on satisfied criteria while the task is still active", () => {
    renderList([PENDING, DONE], { taskStatus: "ready" });

    const editButtons = screen.getAllByRole("button", { name: /^edit$/i });
    expect(editButtons).toHaveLength(1);

    const editDone = screen.getByRole("button", { name: /edit \(locked/i });
    expect(editDone).toBeDefined();
    expect(editDone).toBeDisabled();
    expect(editDone).toHaveAttribute(
      "title",
      expect.stringMatching(/already marked done/i),
    );
  });

  it("allows editing satisfied criteria when the task is done", () => {
    renderList([DONE], { taskStatus: "done" });

    const editDone = screen.getByRole("button", { name: /^edit$/i });
    expect(editDone).toBeEnabled();
  });

  it("locks edit and remove while the task is running", () => {
    renderList([PENDING, DONE], { taskStatus: "running", criteriaLocked: true });

    expect(screen.getAllByRole("button", { name: /edit \(locked/i })).toHaveLength(2);
    expect(screen.getAllByRole("button", { name: /remove.*locked/i })).toHaveLength(2);
  });

  it("does not call onOpenEditCriterionModal when the locked Edit button is clicked", async () => {
    const user = userEvent.setup();
    const { onOpenEditCriterionModal } = renderList([DONE], { taskStatus: "ready" });

    const editDone = screen.getByRole("button", { name: /edit \(locked/i });
    // userEvent honors `disabled` and skips the click; we still
    // explicitly assert no handler invocation, since that is the
    // contract this test pins.
    await user.click(editDone);
    expect(onOpenEditCriterionModal).not.toHaveBeenCalled();
  });

  it("locks the Remove button on satisfied criteria while the task is still active", () => {
    // Mirrors the Edit lock: removing a done criterion would orphan
    // the persisted checklist_item_toggled (done=true) audit row and
    // silently cascade away the per-subject completion row, erasing
    // the historical fact that the task ever satisfied this
    // requirement. The backend rejects this with ErrInvalidInput;
    // this test pins the UI half so users can't even attempt the
    // bogus write. (Earlier behaviour kept Remove enabled as an
    // "escape hatch" for wrong auto-completes — that was rejected as
    // an audit-trail violation; the agent's acceptance is the
    // source of truth.)
    renderList([PENDING, DONE], { taskStatus: "ready" });

    const removeButtons = screen.getAllByRole("button", { name: /^remove/i });
    expect(removeButtons).toHaveLength(2);

    const removePending = removeButtons.find(
      (b) => !b.getAttribute("aria-label")?.includes("locked"),
    );
    expect(removePending).toBeDefined();
    expect(removePending).toBeEnabled();

    const removeDone = removeButtons.find((b) =>
      b.getAttribute("aria-label")?.includes("locked"),
    );
    expect(removeDone).toBeDefined();
    expect(removeDone).toBeDisabled();
    expect(removeDone).toHaveAttribute(
      "title",
      expect.stringMatching(/already marked done/i),
    );
  });

  it("does not call onRemoveChecklistItem when the locked Remove button is clicked", async () => {
    const user = userEvent.setup();
    const { onRemoveChecklistItem } = renderList([DONE], { taskStatus: "ready" });

    const removeDone = screen.getByRole("button", { name: /remove.*locked/i });
    await user.click(removeDone);
    expect(onRemoveChecklistItem).not.toHaveBeenCalled();
  });

  // Pins the verification-popup contract: a satisfied criterion with
  // evidence must NOT inline the payload on the row (older builds used
  // `<details>` disclosures which ballooned the criterion row). It must
  // instead expose a "View verification" trigger that opens a dialog
  // with the evidence. This guards against accidental regression to
  // the inline disclosure pattern.
  it("does not inline evidence on a satisfied row", () => {
    renderList([DONE_VERIFIED]);

    expect(
      screen.queryByText(/CODEBASE_TOUR\.md names Go 1\.25\+/),
    ).not.toBeInTheDocument();
    expect(
      screen.getByRole("button", { name: /view verification details for:/i }),
    ).toBeInTheDocument();
  });

  it("opens the verification modal with evidence on click", async () => {
    const user = userEvent.setup();
    renderList([DONE_VERIFIED]);

    await user.click(
      screen.getByRole("button", { name: /view verification details for:/i }),
    );

    const dialog = screen.getByRole("dialog");
    expect(dialog).toBeInTheDocument();
    expect(
      screen.getByRole("heading", { name: /evidence/i, level: 3 }),
    ).toBeInTheDocument();
    expect(
      screen.queryByRole("heading", { name: /verifier reasoning/i, level: 3 }),
    ).not.toBeInTheDocument();
    expect(
      screen.getByText(/CODEBASE_TOUR\.md names Go 1\.25\+/),
    ).toBeInTheDocument();
  });

  it("closes the verification modal when Close is clicked", async () => {
    const user = userEvent.setup();
    renderList([DONE_VERIFIED]);

    await user.click(
      screen.getByRole("button", { name: /view verification details for:/i }),
    );
    expect(screen.getByRole("dialog")).toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: /^close$/i }));
    expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
  });

  it("does not expose a View verification trigger when no evidence exists", () => {
    // DONE has no evidence. Reasoning-only leftovers from the old
    // verifier path also must not open an empty modal.
    renderList([DONE]);
    expect(
      screen.queryByRole("button", { name: /view verification/i }),
    ).not.toBeInTheDocument();
  });

  it("does not expose a View verification trigger for reasoning-only leftovers", () => {
    renderList([
      {
        ...DONE,
        verifier_reasoning:
          "accepted execute claim (agent self-checked verify commands)",
      },
    ]);
    expect(
      screen.queryByRole("button", { name: /view verification/i }),
    ).not.toBeInTheDocument();
  });

  it("opens edit when anywhere on a pending criterion row is clicked", async () => {
    const user = userEvent.setup();
    const { onOpenEditCriterionModal } = renderList([PENDING]);

    await user.click(
      screen.getByText("Hello World is written inside the 123.md file"),
    );

    expect(onOpenEditCriterionModal).toHaveBeenCalledWith(
      PENDING.id,
      PENDING.text,
      undefined,
    );
  });

  it("opens edit when anywhere on a done criterion row is clicked on a done task", async () => {
    const user = userEvent.setup();
    const { onOpenEditCriterionModal } = renderList([DONE], { taskStatus: "done" });

    await user.click(screen.getByText("123.md file created"));

    expect(onOpenEditCriterionModal).toHaveBeenCalledWith(
      DONE.id,
      DONE.text,
      undefined,
    );
  });

  it("opens view-only when a satisfied row is clicked while the task is still active", async () => {
    const user = userEvent.setup();
    const { onOpenEditCriterionModal } = renderList([DONE], { taskStatus: "ready" });

    await user.click(screen.getByText("123.md file created"));

    expect(onOpenEditCriterionModal).toHaveBeenCalledWith(
      DONE.id,
      DONE.text,
      undefined,
    );
  });

  it("does not call onOpenEditCriterionModal when View verification is clicked on a done row", async () => {
    const user = userEvent.setup();
    const { onOpenEditCriterionModal } = renderList([DONE_VERIFIED]);

    await user.click(
      screen.getByRole("button", { name: /view verification details for:/i }),
    );

    expect(onOpenEditCriterionModal).not.toHaveBeenCalled();
  });

  it("shows verify command badge beside row actions with accessible label", () => {
    renderList([
      {
        ...PENDING,
        verify_commands: [{ command: "go test ./...", expected_outcome: "pass" }],
      },
    ]);

    expect(
      screen.getByLabelText(/1 verify command for the execute agent/i),
    ).toHaveTextContent(
      "1 command",
    );
    expect(screen.getByText("Hello World is written inside the 123.md file")).toBeInTheDocument();
  });
});
