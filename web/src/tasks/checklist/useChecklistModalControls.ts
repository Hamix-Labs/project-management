import { useCallback, useEffect, useRef, useState } from "react";
import type { ChecklistVerifyCommandInput } from "@/types";

function resetChecklistAddFormState(
  setNewChecklistText: (text: string) => void,
  setNewChecklistVerifyCommands: (commands: ChecklistVerifyCommandInput[]) => void,
): void {
  setNewChecklistText("");
  setNewChecklistVerifyCommands([]);
}

function resetEditCriterionFormState(
  setEditingChecklistItemId: (id: string | null) => void,
  setEditChecklistText: (text: string) => void,
  setEditChecklistOriginalText: (text: string) => void,
  setEditChecklistVerifyCommands: (commands: ChecklistVerifyCommandInput[]) => void,
  setEditChecklistOriginalVerifyCommands: (commands: ChecklistVerifyCommandInput[]) => void,
): void {
  setEditingChecklistItemId(null);
  setEditChecklistText("");
  setEditChecklistOriginalText("");
  setEditChecklistVerifyCommands([]);
  setEditChecklistOriginalVerifyCommands([]);
}

export function useChecklistModalControls(taskId: string) {
  const [checklistModalOpen, setChecklistModalOpen] = useState(false);
  const [newChecklistText, setNewChecklistText] = useState("");
  const [newChecklistVerifyCommands, setNewChecklistVerifyCommands] = useState<
    ChecklistVerifyCommandInput[]
  >([]);
  const [editCriterionModalOpen, setEditCriterionModalOpen] = useState(false);
  const [editingChecklistItemId, setEditingChecklistItemId] = useState<string | null>(
    null,
  );
  const [editChecklistText, setEditChecklistText] = useState("");
  const [editChecklistVerifyCommands, setEditChecklistVerifyCommands] = useState<
    ChecklistVerifyCommandInput[]
  >([]);
  const [editChecklistOriginalText, setEditChecklistOriginalText] = useState("");
  const [editChecklistOriginalVerifyCommands, setEditChecklistOriginalVerifyCommands] =
    useState<ChecklistVerifyCommandInput[]>([]);
  const addSubmissionTokenRef = useRef(0);
  const editingChecklistItemIdRef = useRef<string | null>(null);

  useEffect(() => {
    editingChecklistItemIdRef.current = editingChecklistItemId;
  }, [editingChecklistItemId]);

  useEffect(() => {
    setChecklistModalOpen(false);
    resetChecklistAddFormState(setNewChecklistText, setNewChecklistVerifyCommands);
    setEditCriterionModalOpen(false);
    resetEditCriterionFormState(
      setEditingChecklistItemId,
      setEditChecklistText,
      setEditChecklistOriginalText,
      setEditChecklistVerifyCommands,
      setEditChecklistOriginalVerifyCommands,
    );
    addSubmissionTokenRef.current += 1;
  }, [taskId]);

  const closeChecklistModal = useCallback(() => {
    addSubmissionTokenRef.current += 1;
    setChecklistModalOpen(false);
    resetChecklistAddFormState(setNewChecklistText, setNewChecklistVerifyCommands);
  }, []);

  const closeEditCriterionModal = useCallback(() => {
    setEditCriterionModalOpen(false);
    resetEditCriterionFormState(
      setEditingChecklistItemId,
      setEditChecklistText,
      setEditChecklistOriginalText,
      setEditChecklistVerifyCommands,
      setEditChecklistOriginalVerifyCommands,
    );
  }, []);

  const openChecklistModal = useCallback(() => {
    addSubmissionTokenRef.current += 1;
    resetChecklistAddFormState(setNewChecklistText, setNewChecklistVerifyCommands);
    setChecklistModalOpen(true);
    setEditCriterionModalOpen(false);
    resetEditCriterionFormState(
      setEditingChecklistItemId,
      setEditChecklistText,
      setEditChecklistOriginalText,
      setEditChecklistVerifyCommands,
      setEditChecklistOriginalVerifyCommands,
    );
  }, []);

  const openEditCriterionModal = useCallback(
    (itemId: string, text: string, verifyCommands: ChecklistVerifyCommandInput[] = []) => {
      addSubmissionTokenRef.current += 1;
      setEditingChecklistItemId(itemId);
      setEditChecklistText(text);
      setEditChecklistOriginalText(text);
      const cmds = verifyCommands ?? [];
      setEditChecklistVerifyCommands(cmds);
      setEditChecklistOriginalVerifyCommands(cmds);
      setEditCriterionModalOpen(true);
      setChecklistModalOpen(false);
      resetChecklistAddFormState(setNewChecklistText, setNewChecklistVerifyCommands);
    },
    [],
  );

  return {
    checklistModalOpen,
    setChecklistModalOpen,
    newChecklistText,
    setNewChecklistText,
    newChecklistVerifyCommands,
    setNewChecklistVerifyCommands,
    editCriterionModalOpen,
    editingChecklistItemId,
    editChecklistText,
    setEditChecklistText,
    editChecklistVerifyCommands,
    setEditChecklistVerifyCommands,
    editChecklistOriginalText,
    editChecklistOriginalVerifyCommands,
    addSubmissionTokenRef,
    editingChecklistItemIdRef,
    closeChecklistModal,
    closeEditCriterionModal,
    openChecklistModal,
    openEditCriterionModal,
  };
}
