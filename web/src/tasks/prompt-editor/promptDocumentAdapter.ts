import {
  getTask,
  getTaskDraft,
  getTaskTemplate,
  patchTask,
  patchTaskTemplate,
  saveTaskDraft,
} from "@/api";
import type {
  PromptDocumentAdapter,
  PromptDocumentRef,
  PromptDocumentSnapshot,
  PromptSourceKind,
} from "./types";
import {
  PROMPT_EPHEMERAL_PREFIX,
  type PromptEditorLaunchContext,
} from "./types";

function ephemeralStorageKey(id: string): string {
  return `${PROMPT_EPHEMERAL_PREFIX}${id}`;
}

export function readEphemeralPrompt(id: string): PromptDocumentSnapshot {
  try {
    const raw = sessionStorage.getItem(ephemeralStorageKey(id));
    if (!raw) return { html: "" };
    const parsed = JSON.parse(raw) as {
      html?: string;
      worktreeId?: string;
      repositoryId?: string;
      name?: string;
    };
    return {
      html: typeof parsed.html === "string" ? parsed.html : "",
      worktreeId:
        typeof parsed.worktreeId === "string" ? parsed.worktreeId : undefined,
      repositoryId:
        typeof parsed.repositoryId === "string"
          ? parsed.repositoryId
          : undefined,
      name: typeof parsed.name === "string" ? parsed.name : undefined,
    };
  } catch {
    return { html: "" };
  }
}

export function writeEphemeralPrompt(
  id: string,
  snap: PromptDocumentSnapshot,
): void {
  sessionStorage.setItem(
    ephemeralStorageKey(id),
    JSON.stringify({
      html: snap.html,
      worktreeId: snap.worktreeId,
      repositoryId: snap.repositoryId,
      name: snap.name,
    }),
  );
}

export function createPromptDocumentAdapter(
  ref: PromptDocumentRef,
  launch?: PromptEditorLaunchContext,
): PromptDocumentAdapter {
  switch (ref.kind) {
    case "draft":
      return {
        async load(signal) {
          const draft = await getTaskDraft(ref.id, { signal });
          return {
            html: draft.payload.initial_prompt ?? "",
            name: draft.name,
            worktreeId: launch?.worktreeId,
            repositoryId: launch?.repositoryId,
          };
        },
        async save(html) {
          const draft = await getTaskDraft(ref.id);
          await saveTaskDraft({
            id: draft.id,
            name: draft.name,
            payload: {
              ...draft.payload,
              initial_prompt: html,
            },
          });
        },
      };
    case "task":
      return {
        async load(signal) {
          const task = await getTask(ref.id, { signal });
          return {
            html: task.initial_prompt ?? "",
            name: task.title,
            worktreeId: launch?.worktreeId ?? task.worktree_id ?? undefined,
            repositoryId:
              launch?.repositoryId ?? task.repository_id ?? undefined,
          };
        },
        async save(html) {
          await patchTask(ref.id, { initial_prompt: html });
        },
      };
    case "template":
      return {
        async load(signal) {
          if (ref.id === "new") {
            return {
              html: launch ? readEphemeralPrompt(`template-new`).html : "",
              name: "New template",
              worktreeId: launch?.worktreeId,
              repositoryId: launch?.repositoryId,
            };
          }
          const tpl = await getTaskTemplate(ref.id, { signal });
          return {
            html: tpl.payload.initial_prompt ?? "",
            name: tpl.name,
            worktreeId: launch?.worktreeId,
            repositoryId: launch?.repositoryId,
          };
        },
        async save(html) {
          if (ref.id === "new") {
            writeEphemeralPrompt("template-new", {
              html,
              worktreeId: launch?.worktreeId,
              repositoryId: launch?.repositoryId,
            });
            return;
          }
          const tpl = await getTaskTemplate(ref.id);
          await patchTaskTemplate(ref.id, {
            name: tpl.name,
            payload: {
              ...tpl.payload,
              initial_prompt: html,
            },
          });
        },
      };
    case "ephemeral":
      return {
        async load() {
          const snap = readEphemeralPrompt(ref.id);
          return {
            ...snap,
            worktreeId: launch?.worktreeId ?? snap.worktreeId,
            repositoryId: launch?.repositoryId ?? snap.repositoryId,
          };
        },
        async save(html) {
          const prev = readEphemeralPrompt(ref.id);
          writeEphemeralPrompt(ref.id, {
            html,
            worktreeId: launch?.worktreeId ?? prev.worktreeId,
            repositoryId: launch?.repositoryId ?? prev.repositoryId,
            name: prev.name,
          });
        },
      };
    default: {
      const _exhaustive: never = ref.kind;
      throw new Error(`Unknown prompt source: ${_exhaustive}`);
    }
  }
}

export function isPromptSourceKind(value: string): value is PromptSourceKind {
  return (
    value === "draft" ||
    value === "task" ||
    value === "template" ||
    value === "ephemeral"
  );
}
