import { useCallback, useState, type FormEvent } from "react";
import type { ChecklistItemDraft } from "@/types";
import { normalizeVerifyCommands } from "@/tasks/task-compose/checklistRequirement";

type Args = {
  onAppendChecklistCriterion: (item: ChecklistItemDraft) => void;
  onUpdateChecklistRow: (index: number, item: ChecklistItemDraft) => void;
};

export function useChecklistCriterionModal({
  onAppendChecklistCriterion,
  onUpdateChecklistRow,
}: Args) {
  const [criterionModalOpen, setCriterionModalOpen] = useState(false);
  const [criterionModalText, setCriterionModalText] = useState("");
  const [criterionModalCommands, setCriterionModalCommands] = useState<
    ChecklistItemDraft["verify_commands"]
  >([]);
  const [criterionEditIndex, setCriterionEditIndex] = useState<number | null>(
    null,
  );

  const openCriterionModal = useCallback(() => {
    setCriterionEditIndex(null);
    setCriterionModalText("");
    setCriterionModalCommands([]);
    setCriterionModalOpen(true);
  }, []);

  const openEditCriterionModal = useCallback(
    (index: number, item: ChecklistItemDraft) => {
      setCriterionEditIndex(index);
      setCriterionModalText(item.text);
      setCriterionModalCommands(item.verify_commands ?? []);
      setCriterionModalOpen(true);
    },
    [],
  );

  const closeCriterionModal = useCallback(() => {
    setCriterionModalOpen(false);
    setCriterionEditIndex(null);
    setCriterionModalText("");
    setCriterionModalCommands([]);
  }, []);

  const submitCriterionModal = (e: FormEvent) => {
    e.preventDefault();
    e.stopPropagation();
    const t = criterionModalText.trim();
    if (!t) return;
    const item: ChecklistItemDraft = {
      text: t,
      verify_commands: normalizeVerifyCommands(criterionModalCommands ?? []),
    };
    if (criterionEditIndex === null) {
      onAppendChecklistCriterion(item);
    } else {
      onUpdateChecklistRow(criterionEditIndex, item);
    }
    closeCriterionModal();
  };

  return {
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
  };
}
