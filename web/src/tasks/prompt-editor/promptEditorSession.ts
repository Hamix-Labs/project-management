import type {
  PromptEditorLaunchContext,
  PromptEditorReturnPayload,
} from "./types";
import { PROMPT_EDITOR_LAUNCH_KEY, PROMPT_RETURN_KEY } from "./types";

export function writePromptEditorLaunch(
  ctx: PromptEditorLaunchContext,
): void {
  sessionStorage.setItem(PROMPT_EDITOR_LAUNCH_KEY, JSON.stringify(ctx));
}

export function readPromptEditorLaunch(): PromptEditorLaunchContext | null {
  try {
    const raw = sessionStorage.getItem(PROMPT_EDITOR_LAUNCH_KEY);
    if (!raw) return null;
    return JSON.parse(raw) as PromptEditorLaunchContext;
  } catch {
    return null;
  }
}

export function clearPromptEditorLaunch(): void {
  sessionStorage.removeItem(PROMPT_EDITOR_LAUNCH_KEY);
}

export function writePromptEditorReturn(
  payload: PromptEditorReturnPayload,
): void {
  sessionStorage.setItem(PROMPT_RETURN_KEY, JSON.stringify(payload));
}

export function consumePromptEditorReturn(): PromptEditorReturnPayload | null {
  try {
    const raw = sessionStorage.getItem(PROMPT_RETURN_KEY);
    if (!raw) return null;
    sessionStorage.removeItem(PROMPT_RETURN_KEY);
    return JSON.parse(raw) as PromptEditorReturnPayload;
  } catch {
    sessionStorage.removeItem(PROMPT_RETURN_KEY);
    return null;
  }
}

export function promptEditorPath(kind: string, id: string): string {
  return `/prompt/${encodeURIComponent(kind)}/${encodeURIComponent(id)}`;
}

export function generateEphemeralPromptId(): string {
  return typeof crypto !== "undefined" && typeof crypto.randomUUID === "function"
    ? crypto.randomUUID()
    : `ephemeral-${Date.now().toString(36)}`;
}
