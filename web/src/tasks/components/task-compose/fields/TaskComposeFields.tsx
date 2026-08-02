import type { ReactNode } from "react";
import type { PriorityChoice, ChecklistItemDraft } from "@/types";
import { ChecklistCriterionModal } from "../modals/ChecklistCriterionModal";
import { TaskComposeChecklistFields } from "./TaskComposeChecklistFields";
import { TaskComposePromptField } from "./TaskComposePromptField";
import { TaskComposeTitlePriorityRow } from "./TaskComposeTitlePriorityRow";
import { useChecklistCriterionModal } from "../hooks/useChecklistCriterionModal";

export type TaskComposeFieldsProps = {
  /** Prefix for stable `id`s, e.g. `task-new` → `task-new-title`. */
  idsPrefix: string;
  title: string;
  prompt: string;
  priority: PriorityChoice;
  checklistItems: ChecklistItemDraft[];
  /** When true, the done-criteria block is omitted (e.g. subtask inherits a parent checklist). */
  hideChecklist?: boolean;
  /** When `required`, at least one done criterion is required on create. */
  checklistRequirement?: "optional" | "required";
  /** When true, checklist add/edit/remove controls are disabled. */
  checklistDisabled?: boolean;
  disabled: boolean;
  onTitleChange: (v: string) => void;
  onOpenPromptEditor: () => void;
  onPriorityChange: (p: PriorityChoice) => void;
  onAppendChecklistCriterion: (item: ChecklistItemDraft) => void;
  onUpdateChecklistRow: (index: number, item: ChecklistItemDraft) => void;
  onRemoveChecklistRow: (index: number) => void;
  /** Rendered between the title row and the prompt editor (e.g. git binding). */
  betweenTitleAndPrompt?: ReactNode;
};

export function TaskComposeFields({
  idsPrefix,
  title,
  prompt,
  priority,
  checklistItems,
  hideChecklist = false,
  checklistRequirement = "optional",
  checklistDisabled = false,
  disabled,
  onTitleChange,
  onOpenPromptEditor,
  onPriorityChange,
  onAppendChecklistCriterion,
  onUpdateChecklistRow,
  onRemoveChecklistRow,
  betweenTitleAndPrompt,
}: TaskComposeFieldsProps) {
  const checklistHeadingId = `${idsPrefix}-checklist-heading`;

  const {
    criterionModalOpen,
    criterionModalText,
    criterionModalCommands,
    criterionEditIndex,
    setCriterionModalText,
    setCriterionModalCommands,
    openCriterionModal,
    openEditCriterionModal,
    closeCriterionModal,
    submitCriterionModal,
  } = useChecklistCriterionModal({
    onAppendChecklistCriterion,
    onUpdateChecklistRow,
  });

  return (
    <>
      <TaskComposeTitlePriorityRow
        idsPrefix={idsPrefix}
        title={title}
        priority={priority}
        disabled={disabled}
        onTitleChange={onTitleChange}
        onPriorityChange={onPriorityChange}
      />

      {betweenTitleAndPrompt}

      <TaskComposePromptField
        idsPrefix={idsPrefix}
        prompt={prompt}
        disabled={disabled}
        onOpenPromptEditor={onOpenPromptEditor}
      />

      {!hideChecklist ? (
        <TaskComposeChecklistFields
          checklistHeadingId={checklistHeadingId}
          checklistItems={checklistItems}
          checklistRequirement={checklistRequirement}
          disabled={disabled || checklistDisabled}
          onOpenNewCriterion={openCriterionModal}
          onOpenEditCriterion={openEditCriterionModal}
          onRemoveRow={onRemoveChecklistRow}
        />
      ) : null}

      {criterionModalOpen ? (
        <ChecklistCriterionModal
          mode={criterionEditIndex === null ? "add" : "edit"}
          pending={false}
          saving={false}
          onClose={closeCriterionModal}
          text={criterionModalText}
          onTextChange={setCriterionModalText}
          verifyCommands={criterionModalCommands ?? []}
          onVerifyCommandsChange={setCriterionModalCommands}
          onSubmit={submitCriterionModal}
          modalStack="nested"
          lockBodyScroll={false}
        />
      ) : null}
    </>
  );
}
