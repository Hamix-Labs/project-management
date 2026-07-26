import { useCallback, useEffect, useRef, useState, type FormEvent } from "react";
import type { Priority, Task } from "@/types";
import { canEditTask } from "../task-display/canEditTask";
import { validateTagsCsv } from "../create/taskTagValidation";
import { buildPatchMutationInput } from "../mutations/buildPatchMutationInput";
import { useTaskPatchFlow } from "./useTaskPatchFlow";

export type UseTaskEditFlowOptions = {
  editingTaskId: string | null;
  closeCreateModal: () => void;
  beginEditSession: (t: Task) => void | Promise<void>;
  newTitle: string;
  newPrompt: string;
  newPriority: string;
  newProjectID: string;
  newTagsCsv: string;
  newMilestone: string;
  newTaskCursorModel: string;
  newTaskVerifyChatMode: string;
  newSchedule: string | null;
  composeStatus: import("@/types").Status;
  createModalOpen: boolean;
};

/**
 * Edit + change-model modal state and PATCH submit. Composed by `useTasksApp`.
 */
export function useTaskEditFlow(opts: UseTaskEditFlowOptions) {
  const editingTaskIdRef = useRef<string | null>(null);
  editingTaskIdRef.current = opts.editingTaskId;

  const [changeModelTask, setChangeModelTask] = useState<Task | null>(null);
  const [changeModelDraft, setChangeModelDraft] = useState("");
  const [editTitleRequiredError, setEditTitleRequiredError] = useState<
    string | null
  >(null);

  const {
    patchTask: runPatch,
    patchPending,
    patchError,
    resetError: resetPatchError,
  } = useTaskPatchFlow({
    onPatched: (patchedId) => {
      if (editingTaskIdRef.current === patchedId) {
        opts.closeCreateModal();
      }
      setChangeModelTask((prev) => (prev?.id === patchedId ? null : prev));
    },
  });

  useEffect(() => {
    if (!opts.createModalOpen && !changeModelTask) resetPatchError();
  }, [opts.createModalOpen, changeModelTask, resetPatchError]);

  useEffect(() => {
    if (editTitleRequiredError === "Title is required." && opts.newTitle.trim()) {
      setEditTitleRequiredError(null);
    }
  }, [opts.newTitle, editTitleRequiredError]);

  useEffect(() => {
    if (
      editTitleRequiredError &&
      editTitleRequiredError.startsWith("Tag ") &&
      validateTagsCsv(opts.newTagsCsv) === null
    ) {
      setEditTitleRequiredError(null);
    }
  }, [opts.newTagsCsv, editTitleRequiredError]);

  const openEdit = useCallback(
    (t: Task) => {
      if (!canEditTask(t.status)) {
        return;
      }
      setChangeModelTask(null);
      setEditTitleRequiredError(null);
      void opts.beginEditSession(t);
    },
    [opts.beginEditSession],
  );

  const closeEdit = useCallback(() => {
    opts.closeCreateModal();
    setEditTitleRequiredError(null);
  }, [opts.closeCreateModal]);

  const openChangeModel = useCallback(
    (t: Task) => {
      if (opts.editingTaskId) {
        opts.closeCreateModal();
      }
      setEditTitleRequiredError(null);
      setChangeModelTask(t);
      setChangeModelDraft(t.cursor_model ?? "");
    },
    [opts.editingTaskId, opts.closeCreateModal],
  );

  const closeChangeModel = useCallback(() => {
    setChangeModelTask(null);
  }, []);

  const submitChangeModel = useCallback(
    (e: FormEvent) => {
      e.preventDefault();
      const t = changeModelTask;
      if (!t) return;
      runPatch({
        id: t.id,
        title: t.title.trim(),
        initial_prompt: t.initial_prompt,
        status: t.status,
        priority: t.priority,
        project_id: t.project_id ?? null,
        tags: t.tags ?? [],
        milestone: t.milestone ?? null,
        cursor_model: changeModelDraft.trim(),
      });
    },
    [changeModelTask, changeModelDraft, runPatch],
  );

  const submitEdit = useCallback(
    (e: FormEvent) => {
      e.preventDefault();
      if (!opts.editingTaskId || !opts.newPriority) return;
      if (!opts.newTitle.trim()) {
        setEditTitleRequiredError("Title is required.");
        return;
      }
      const tagsError = validateTagsCsv(opts.newTagsCsv);
      if (tagsError) {
        setEditTitleRequiredError(tagsError);
        return;
      }
      setEditTitleRequiredError(null);
      runPatch(
        buildPatchMutationInput({
          id: opts.editingTaskId,
          title: opts.newTitle,
          initial_prompt: opts.newPrompt,
          status: opts.composeStatus,
          priority: opts.newPriority as Priority,
          project_id: opts.newProjectID.trim() || null,
          tagsCsv: opts.newTagsCsv,
          milestone: opts.newMilestone,
          cursor_model: opts.newTaskCursorModel,
          verify_chat_mode: opts.newTaskVerifyChatMode,
          pickup_not_before: opts.newSchedule,
        }),
      );
    },
    [
      opts.editingTaskId,
      opts.newPriority,
      opts.newTitle,
      opts.newPrompt,
      opts.composeStatus,
      opts.newProjectID,
      opts.newTagsCsv,
      opts.newMilestone,
      opts.newTaskCursorModel,
      opts.newTaskVerifyChatMode,
      opts.newSchedule,
      runPatch,
    ],
  );

  return {
    changeModelTask,
    changeModelDraft,
    setChangeModelDraft,
    openChangeModel,
    closeChangeModel,
    submitChangeModel,
    openEdit,
    closeEdit,
    submitEdit,
    editTitleRequiredError,
    patchPending,
    patchError,
    resetPatchError,
  };
}
