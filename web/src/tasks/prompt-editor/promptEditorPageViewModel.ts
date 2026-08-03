import type { PromptEditorSaveStatusKind } from "@/components/prompt-editor/chrome";
import {
  formatRelativeEdited,
  wordCountFromHtml,
} from "./promptEditorPageMeta";
import type { PromptEditorLoadStatus } from "./usePromptEditorDocumentLoad";
import type { PromptEditorSessionError } from "./promptEditorSessionError";

export function pickLoadError(
  status: PromptEditorLoadStatus,
  sessionError: PromptEditorSessionError | null,
): PromptEditorSessionError | null {
  return status === "error" && sessionError?.phase === "load"
    ? sessionError
    : null;
}

export function pickSaveError(
  sessionError: PromptEditorSessionError | null,
): PromptEditorSessionError | null {
  return sessionError?.phase === "save" || sessionError?.phase === "leave"
    ? sessionError
    : null;
}

export function deriveSaveStatus(input: {
  saveError: PromptEditorSessionError | null;
  saving: boolean;
  leavePending: boolean;
  dirty: boolean;
}): PromptEditorSaveStatusKind {
  if (input.saveError) return "error";
  if (input.saving || input.leavePending) return "saving";
  if (input.dirty) return "unsaved";
  return "saved";
}

export function deriveEditedLabel(
  status: PromptEditorLoadStatus,
  ready: boolean,
  lastSavedAt: number | null,
): string {
  if (ready) return formatRelativeEdited(lastSavedAt);
  if (status === "loading") return "Loading…";
  return "Not edited yet";
}

export function deriveWordCountLabel(ready: boolean, html: string): string {
  if (!ready) return "—";
  const words = wordCountFromHtml(html);
  return words === 0 ? "0 words" : `~${words} words`;
}
