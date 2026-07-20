import {
  useCallback,
  useReducer,
  useRef,
  useState,
  type Dispatch,
  type SetStateAction,
} from "react";
import { listChecklist } from "@/api";
import { DEFAULT_NEW_TASK_STATUS, type ChecklistItemDraft, type Status, type Task } from "@/types";
import type { ComposeOperation, ComposeTarget, CreateModalPrefill } from "../types";
import {
  deriveCreateUiFlags,
  initialCreateUiPhase,
  reduceCreateUiPhase,
  type CreateUiPhase,
  type CreateUiPhaseEvent,
} from "../createUiPhase";

function createUiPhaseReducer(
  phase: CreateUiPhase,
  event: CreateUiPhaseEvent,
): CreateUiPhase {
  return reduceCreateUiPhase(phase, event);
}

export function useTaskCreateModalState(
  resetFormFields: () => void,
  populateFromTask: (t: Task) => void,
  setNewChecklistItems: Dispatch<SetStateAction<ChecklistItemDraft[]>>,
  setNewProjectID: (id: string) => void,
) {
  const createModalPrefillRef = useRef<CreateModalPrefill | null>(null);
  const [uiPhase, dispatchUiPhase] = useReducer(
    createUiPhaseReducer,
    undefined,
    initialCreateUiPhase,
  );
  const uiFlags = deriveCreateUiFlags(uiPhase);

  const [editingTaskRunner, setEditingTaskRunner] = useState("");
  const [composeStatus, setComposeStatus] = useState<Status>(DEFAULT_NEW_TASK_STATUS);
  const [createModalAssignmentLocked, setCreateModalAssignmentLocked] = useState(false);
  const [createEntryDraftErrorHint, setCreateEntryDraftErrorHint] = useState<
    string | null
  >(null);

  /** Monotonic session id so stale checklist fetches cannot stamp the next edit. */
  const editSessionGenerationRef = useRef(0);
  const editChecklistAbortRef = useRef<AbortController | null>(null);
  /** Monotonic entry id so late ensureRepositoriesRegistered cannot open UI after supersede. */
  const entryRequestIdRef = useRef(0);

  const resetNewTaskForm = useCallback(() => {
    resetFormFields();
    setCreateModalAssignmentLocked(false);
    setEditingTaskRunner("");
    setComposeStatus(DEFAULT_NEW_TASK_STATUS);
  }, [resetFormFields]);

  const applyCreateModalPrefill = useCallback(() => {
    const prefill = createModalPrefillRef.current;
    if (!prefill?.projectID) return;
    setNewProjectID(prefill.projectID);
    setCreateModalAssignmentLocked(prefill.lockProjectAssignment);
    createModalPrefillRef.current = null;
  }, [setNewProjectID]);

  const closeCreateModal = useCallback(() => {
    editSessionGenerationRef.current += 1;
    editChecklistAbortRef.current?.abort();
    editChecklistAbortRef.current = null;
    entryRequestIdRef.current += 1;
    createModalPrefillRef.current = null;
    setCreateEntryDraftErrorHint(null);
    dispatchUiPhase({ type: "close" });
    resetNewTaskForm();
  }, [resetNewTaskForm]);

  const beginEditSession = useCallback(
    async (t: Task) => {
      editChecklistAbortRef.current?.abort();
      const sessionId = ++editSessionGenerationRef.current;
      const abort = new AbortController();
      editChecklistAbortRef.current = abort;

      populateFromTask(t);
      setEditingTaskRunner(t.runner);
      setComposeStatus(t.status);
      setNewChecklistItems([]);
      setCreateEntryDraftErrorHint(null);
      dispatchUiPhase({
        type: "openCompose",
        target: "task",
        operation: "edit",
        editingTaskId: t.id,
        editingTemplateId: null,
      });
      try {
        const { items } = await listChecklist(t.id, { signal: abort.signal });
        if (editSessionGenerationRef.current !== sessionId) {
          return;
        }
        setNewChecklistItems(
          items.map((item) => ({
            text: item.text,
            verify_commands: item.verify_commands,
          })),
        );
      } catch {
        // Checklist is display-only in edit; leave empty on fetch failure / abort.
      }
    },
    [populateFromTask, setNewChecklistItems],
  );

  const setDraftPickerOpen = useCallback((open: boolean) => {
    if (open) {
      dispatchUiPhase({ type: "showDraftPicker" });
      return;
    }
    if (uiPhase.kind === "draftPicker") {
      dispatchUiPhase({ type: "close" });
    }
  }, [uiPhase.kind]);

  const setCreateModalOpen = useCallback(
    (open: boolean) => {
      if (open) {
        dispatchUiPhase({
          type: "openCompose",
          target: uiFlags.composeTarget,
          operation: uiFlags.composeOperation,
          editingTaskId: uiFlags.editingTaskId,
          editingTemplateId: uiFlags.editingTemplateId,
        });
        return;
      }
      dispatchUiPhase({ type: "close" });
      resetNewTaskForm();
    },
    [
      uiFlags.composeTarget,
      uiFlags.composeOperation,
      uiFlags.editingTaskId,
      uiFlags.editingTemplateId,
      resetNewTaskForm,
    ],
  );

  const setRepositorySetupPromptOpen = useCallback((open: boolean) => {
    if (open) {
      dispatchUiPhase({ type: "showRepositorySetup" });
      return;
    }
    if (uiPhase.kind === "repositorySetup") {
      dispatchUiPhase({ type: "close" });
    }
  }, [uiPhase.kind]);

  const setComposeTarget = useCallback(
    (target: ComposeTarget) => {
      if (uiPhase.kind !== "compose") return;
      dispatchUiPhase({
        type: "openCompose",
        target,
        operation: uiPhase.operation,
        editingTaskId: uiPhase.editingTaskId,
        editingTemplateId: uiPhase.editingTemplateId,
      });
    },
    [uiPhase],
  );

  const setComposeOperation = useCallback(
    (operation: ComposeOperation) => {
      if (uiPhase.kind !== "compose") return;
      dispatchUiPhase({
        type: "openCompose",
        target: uiPhase.target,
        operation,
        editingTaskId: uiPhase.editingTaskId,
        editingTemplateId: uiPhase.editingTemplateId,
      });
    },
    [uiPhase],
  );

  const setEditingTemplateId = useCallback(
    (id: string | null) => {
      if (uiPhase.kind !== "compose") return;
      dispatchUiPhase({
        type: "openCompose",
        target: uiPhase.target,
        operation: uiPhase.operation,
        editingTaskId: uiPhase.editingTaskId,
        editingTemplateId: id,
      });
    },
    [uiPhase],
  );

  /** Begin a new async entry attempt; returns the generation to check after await. */
  const beginEntryRequest = useCallback(() => {
    entryRequestIdRef.current += 1;
    return entryRequestIdRef.current;
  }, []);

  const isEntryRequestCurrent = useCallback((requestId: number) => {
    return entryRequestIdRef.current === requestId;
  }, []);

  const openComposePhase = useCallback(
    (opts: {
      target: ComposeTarget;
      operation: ComposeOperation;
      editingTaskId?: string | null;
      editingTemplateId?: string | null;
    }) => {
      dispatchUiPhase({
        type: "openCompose",
        target: opts.target,
        operation: opts.operation,
        editingTaskId: opts.editingTaskId ?? null,
        editingTemplateId: opts.editingTemplateId ?? null,
      });
    },
    [],
  );

  const showDraftPickerPhase = useCallback(() => {
    dispatchUiPhase({ type: "showDraftPicker" });
  }, []);

  const showRepositorySetupPhase = useCallback(() => {
    dispatchUiPhase({ type: "showRepositorySetup" });
  }, []);

  return {
    createModalPrefillRef,
    draftPickerOpen: uiFlags.draftPickerOpen,
    setDraftPickerOpen,
    createModalOpen: uiFlags.createModalOpen,
    setCreateModalOpen,
    editingTaskId: uiFlags.editingTaskId,
    editingTemplateId: uiFlags.editingTemplateId,
    setEditingTemplateId,
    composeTarget: uiFlags.composeTarget,
    setComposeTarget,
    composeOperation: uiFlags.composeOperation,
    setComposeOperation,
    editingTaskRunner,
    composeStatus,
    setComposeStatus,
    createModalAssignmentLocked,
    setCreateModalAssignmentLocked,
    createEntryDraftErrorHint,
    setCreateEntryDraftErrorHint,
    repositorySetupPromptOpen: uiFlags.repositorySetupPromptOpen,
    setRepositorySetupPromptOpen,
    applyCreateModalPrefill,
    resetNewTaskForm,
    closeCreateModal,
    beginEditSession,
    beginEntryRequest,
    isEntryRequestCurrent,
    openComposePhase,
    showDraftPickerPhase,
    showRepositorySetupPhase,
    createUiPhase: uiPhase,
  };
}
