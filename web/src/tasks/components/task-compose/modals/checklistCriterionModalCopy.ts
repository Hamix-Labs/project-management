import type { FormEvent } from "react";
import type { ChecklistVerifyCommandInput } from "@/types";
import {
  emptyVerifyCommandRow,
  MAX_VERIFY_COMMANDS_PER_ITEM,
} from "@/tasks/task-compose/checklistRequirement";

export type ChecklistCriterionModalProps = {
  mode: "add" | "edit";
  /** Satisfied criteria open in read-only view — no edits or saves. */
  readOnly?: boolean;
  pending: boolean;
  saving: boolean;
  onClose: () => void;
  text: string;
  onTextChange: (v: string) => void;
  verifyCommands: ChecklistVerifyCommandInput[];
  onVerifyCommandsChange: (cmds: ChecklistVerifyCommandInput[]) => void;
  onSubmit: (e: FormEvent) => void;
  modalStack?: "default" | "nested";
  lockBodyScroll?: boolean;
  dismissibleWhileBusy?: boolean;
  error?: unknown;
  errorFallback?: string;
};

export type ChecklistCriterionModalCopy = {
  titleId: string;
  title: string;
  lead: string;
  busyLabel: string;
  defaultErrorFallback: string;
};

export function verifyCommandsHint(count: number): string {
  if (count === 0) return "Optional";
  if (count === 1) return "1 command";
  return `${count} commands`;
}

export function resolveChecklistCriterionModalCopy(
  mode: "add" | "edit",
  readOnly: boolean,
): ChecklistCriterionModalCopy {
  const busyLabel = mode === "add" ? "Adding criterion…" : "Saving changes…";
  const defaultErrorFallback =
    mode === "add" ? "Could not add criterion." : "Could not save changes.";

  if (readOnly) {
    return {
      titleId: "checklist-criterion-view-modal-title",
      title: "View criterion",
      lead:
        "This criterion is satisfied and locked. You can review the wording and verify commands, but not change them.",
      busyLabel,
      defaultErrorFallback,
    };
  }
  if (mode === "add") {
    return {
      titleId: "checklist-criterion-modal-title",
      title: "New criterion",
      lead: "One clear, testable requirement for done.",
      busyLabel,
      defaultErrorFallback,
    };
  }
  return {
    titleId: "checklist-criterion-edit-modal-title",
    title: "Edit criterion",
    lead: "Update the wording or verification commands.",
    busyLabel,
    defaultErrorFallback,
  };
}

export function handleChecklistCriterionFormSubmit(
  e: FormEvent,
  readOnly: boolean,
  onSubmit: (e: FormEvent) => void,
): void {
  e.stopPropagation();
  if (readOnly) {
    e.preventDefault();
    return;
  }
  onSubmit(e);
}

export type VerifyCommandHandlers = {
  updateCommand: (
    index: number,
    patch: Partial<ChecklistVerifyCommandInput>,
  ) => void;
  addCommandRow: () => void;
  ensureVerifySectionReady: (open: boolean) => void;
  removeCommandRow: (index: number) => void;
  setVerifySectionOpen: (open: boolean) => void;
};

export function createVerifyCommandHandlers(
  verifyCommands: ChecklistVerifyCommandInput[],
  onVerifyCommandsChange: (cmds: ChecklistVerifyCommandInput[]) => void,
  setVerifySectionOpen: (open: boolean) => void,
): VerifyCommandHandlers {
  const updateCommand = (
    index: number,
    patch: Partial<ChecklistVerifyCommandInput>,
  ) => {
    onVerifyCommandsChange(
      verifyCommands.map((row, i) => (i === index ? { ...row, ...patch } : row)),
    );
  };

  const addCommandRow = () => {
    if (verifyCommands.length >= MAX_VERIFY_COMMANDS_PER_ITEM) return;
    setVerifySectionOpen(true);
    onVerifyCommandsChange([...verifyCommands, emptyVerifyCommandRow()]);
  };

  const ensureVerifySectionReady = (open: boolean) => {
    setVerifySectionOpen(open);
    if (open && verifyCommands.length === 0) {
      onVerifyCommandsChange([emptyVerifyCommandRow()]);
    }
  };

  const removeCommandRow = (index: number) => {
    onVerifyCommandsChange(verifyCommands.filter((_, i) => i !== index));
  };

  return {
    updateCommand,
    addCommandRow,
    ensureVerifySectionReady,
    removeCommandRow,
    setVerifySectionOpen,
  };
}

export function checklistCriterionModalSheetClass(readOnly: boolean): string {
  return readOnly
    ? "panel modal-sheet task-checklist-criterion-modal-sheet task-checklist-criterion-modal-sheet--read-only"
    : "panel modal-sheet task-checklist-criterion-modal-sheet";
}
