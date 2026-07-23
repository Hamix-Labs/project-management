import { useState, type FormEvent } from "react";
import type { ChecklistItemDraft } from "@/types";
import { normalizeVerifyCommands } from "@/tasks/task-compose/checklistRequirement";

export function usePolishCriterionDrafts() {
  const [newCriteria, setNewCriteria] = useState<ChecklistItemDraft[]>([]);
  const [modalOpen, setModalOpen] = useState(false);
  const [modalText, setModalText] = useState("");
  const [modalCommands, setModalCommands] = useState<
    NonNullable<ChecklistItemDraft["verify_commands"]>
  >([]);

  function openModal() {
    setModalText("");
    setModalCommands([]);
    setModalOpen(true);
  }

  function closeModal() {
    setModalOpen(false);
    setModalText("");
    setModalCommands([]);
  }

  function submitModal(e: FormEvent) {
    e.preventDefault();
    e.stopPropagation();
    const text = modalText.trim();
    if (!text) return;
    setNewCriteria((prev) => [
      ...prev,
      {
        text,
        verify_commands: normalizeVerifyCommands(modalCommands),
      },
    ]);
    closeModal();
  }

  function removeAt(index: number) {
    setNewCriteria((prev) => prev.filter((_, i) => i !== index));
  }

  return {
    newCriteria,
    modalOpen,
    modalText,
    modalCommands,
    openModal,
    closeModal,
    setModalText,
    setModalCommands,
    submitModal,
    removeAt,
  };
}
