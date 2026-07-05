import { useEffect } from "react";
import type { ChecklistItemDraft } from "@/types";
import { ChecklistCriterionModal } from "../../task-compose/modals/ChecklistCriterionModal";
import { TaskComposeChecklistFields } from "../../task-compose/fields/TaskComposeChecklistFields";
import { useChecklistCriterionModal } from "../../task-compose/hooks/useChecklistCriterionModal";

type Props = {
  checklistHeadingId?: string;
  checklistItems: ChecklistItemDraft[];
  checklistRequirement?: "optional" | "required";
  checklistDisabled?: boolean;
  disabled: boolean;
  onAppendChecklistCriterion: (item: ChecklistItemDraft | string) => void;
  onUpdateChecklistRow: (index: number, item: ChecklistItemDraft) => void;
  onRemoveChecklistRow: (index: number) => void;
  registerOpenNew?: (open: (() => void) | null) => void;
};

export function TaskCreateModalCriteriaFields({
  checklistHeadingId = "task-create-modal-section-criteria",
  checklistItems,
  checklistRequirement = "optional",
  checklistDisabled = false,
  disabled,
  onAppendChecklistCriterion,
  onUpdateChecklistRow,
  onRemoveChecklistRow,
  registerOpenNew,
}: Props) {
  const resolvedChecklistHeadingId = checklistHeadingId;

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
    onAppendChecklistCriterion: (item) => onAppendChecklistCriterion(item),
    onUpdateChecklistRow,
  });

  useEffect(() => {
    registerOpenNew?.(openCriterionModal);
    return () => registerOpenNew?.(null);
  }, [registerOpenNew, openCriterionModal]);

  return (
    <>
      <TaskComposeChecklistFields
        checklistHeadingId={resolvedChecklistHeadingId}
        checklistItems={checklistItems}
        checklistRequirement={checklistRequirement}
        disabled={disabled || checklistDisabled}
        hideSectionHeading
        hideNewCriterionButton
        onOpenNewCriterion={openCriterionModal}
        onOpenEditCriterion={openEditCriterionModal}
        onRemoveRow={onRemoveChecklistRow}
      />

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
