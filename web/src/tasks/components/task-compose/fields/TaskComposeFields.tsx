import type { ReactNode } from "react";
import type { PriorityChoice, ChecklistItemDraft } from "@/types";
import type { RichPromptEditorProjectContextProps } from "../../rich-prompt";
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
  onPromptChange: (v: string) => void;
  onPriorityChange: (p: PriorityChoice) => void;
  onAppendChecklistCriterion: (item: ChecklistItemDraft) => void;
  onUpdateChecklistRow: (index: number, item: ChecklistItemDraft) => void;
  onRemoveChecklistRow: (index: number) => void;
  /** Passed to `RichPromptEditor` as `key` so the editor resets when needed. */
  editorKey: string;
  /**
   * When provided, the prompt editor wires the `#` project context
   * suggestion plugin and renders the read-only REFERENCES block above the
   * editable area. Pass `undefined` for surfaces where project context
   * isn't applicable (e.g. nested subtask drafts that inherit from parent).
   */
  projectContext?: RichPromptEditorProjectContextProps;
  /** Rendered between the title row and the prompt editor (e.g. git binding). */
  betweenTitleAndPrompt?: ReactNode;
  /** When set, @-mentions resolve against this worktree. */
  worktreeId?: string;
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
  onPromptChange,
  onPriorityChange,
  onAppendChecklistCriterion,
  onUpdateChecklistRow,
  onRemoveChecklistRow,
  editorKey,
  projectContext,
  betweenTitleAndPrompt,
  worktreeId,
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
        editorKey={editorKey}
        prompt={prompt}
        disabled={disabled}
        onPromptChange={onPromptChange}
        projectContext={projectContext}
        worktreeId={worktreeId}
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
