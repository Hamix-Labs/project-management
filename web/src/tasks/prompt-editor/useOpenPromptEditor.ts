import type { NavigateFunction } from "react-router-dom";
import { useNavigate } from "react-router-dom";
import { useCallback } from "react";
import { buildDraftSavePayload, computeDraftAutosaveSignature } from "../create/draftPayload";
import type { useTaskCreateFormState } from "../create/hooks/useTaskCreateFormState";
import type { useTaskCreateModalState } from "../create/hooks/useTaskCreateModalState";
import type { useTaskCreateMutations } from "../create/hooks/useTaskCreateMutations";
import { writeEphemeralPrompt } from "./promptDocumentAdapter";
import {
  generateEphemeralPromptId,
  promptEditorPath,
  writePromptEditorLaunch,
} from "./promptEditorSession";
import type { PromptSourceKind } from "./types";

type ComposeOpenInput = {
  form: ReturnType<typeof useTaskCreateFormState>;
  modal: ReturnType<typeof useTaskCreateModalState>;
  mutations: ReturnType<typeof useTaskCreateMutations>;
};

/**
 * Flush compose state and navigate to the full-page Prompt Editor.
 * Takes `navigate` so create-flow hooks do not require a Router (unit tests).
 */
export function createOpenComposePromptEditor(input: ComposeOpenInput) {
  const { form, modal, mutations } = input;

  return async (navigate: NavigateFunction) => {
    if (!modal.createModalOpen) return;

    let kind: PromptSourceKind;
    let id: string;

    if (modal.editingTaskId) {
      kind = "task";
      id = modal.editingTaskId;
    } else if (modal.composeTarget === "template") {
      kind = "template";
      id = modal.editingTemplateId?.trim() || "new";
      if (id === "new") {
        writeEphemeralPrompt("template-new", {
          html: form.newPrompt,
          worktreeId: form.newWorktreeID.trim() || undefined,
        });
      }
    } else {
      kind = "draft";
      id = form.newDraftID;
      if (!id) return;
      const signature = computeDraftAutosaveSignature(form.formFields);
      try {
        await mutations.saveDraftMutation.mutateAsync({
          ...buildDraftSavePayload(form.formFields),
          signature,
        });
      } catch {
        // Still open the editor; page can retry save.
      }
    }

    writePromptEditorLaunch({
      worktreeId: form.newWorktreeID.trim() || undefined,
      returnPath: "/",
      resumeCompose: true,
      seedHtml: form.newPrompt,
      title:
        modal.composeTarget === "template"
          ? "Template prompt"
          : modal.editingTaskId
            ? "Edit prompt"
            : "Initial prompt",
      placeholder: "Write the implementation brief…",
    });

    modal.suspendForPromptEditor();
    navigate(promptEditorPath(kind, id));
  };
}

/** Open ephemeral Prompt Editor for polish instructions. */
export function useOpenPolishPromptEditor() {
  const navigate = useNavigate();

  return useCallback(
    (opts: {
      instructionsHtml: string;
      worktreeId?: string;
      taskId: string;
      returnPath: string;
      ephemeralId?: string;
    }) => {
      const id = opts.ephemeralId ?? generateEphemeralPromptId();
      writeEphemeralPrompt(id, {
        html: opts.instructionsHtml,
        worktreeId: opts.worktreeId,
        name: "Polish instructions",
      });
      writePromptEditorLaunch({
        worktreeId: opts.worktreeId,
        returnPath: opts.returnPath,
        resumePolish: true,
        polishTaskId: opts.taskId,
        title: "Polish instructions",
        placeholder: "Describe what should change in this polish pass…",
      });
      navigate(promptEditorPath("ephemeral", id));
      return id;
    },
    [navigate],
  );
}
