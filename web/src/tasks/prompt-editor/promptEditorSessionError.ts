import { errorMessage } from "@/lib/errorMessage";

export type PromptEditorPhase = "load" | "hydrate" | "save" | "leave";

export type PromptEditorSessionErrorCode =
  | "load_failed"
  | "hydrate_fallback"
  | "save_failed"
  | "leave_save_failed"
  | "rename_failed";

/** First-class session failure — title + detail + phase for actionable UI. */
export type PromptEditorSessionError = {
  phase: PromptEditorPhase;
  title: string;
  detail: string;
  code: PromptEditorSessionErrorCode;
};

export function sessionErrorFromUnknown(
  phase: PromptEditorPhase,
  err: unknown,
  opts: {
    title: string;
    code: PromptEditorSessionErrorCode;
    fallbackDetail: string;
  },
): PromptEditorSessionError {
  return {
    phase,
    title: opts.title,
    detail: errorMessage(err, opts.fallbackDetail),
    code: opts.code,
  };
}

export function loadSessionError(err: unknown): PromptEditorSessionError {
  return sessionErrorFromUnknown("load", err, {
    title: "Couldn't load this prompt",
    code: "load_failed",
    fallbackDetail: "The prompt document could not be loaded.",
  });
}

export function saveSessionError(err: unknown): PromptEditorSessionError {
  return sessionErrorFromUnknown("save", err, {
    title: "Couldn't save",
    code: "save_failed",
    fallbackDetail: "Saving the prompt failed.",
  });
}

export function leaveSaveSessionError(err: unknown): PromptEditorSessionError {
  return sessionErrorFromUnknown("leave", err, {
    title: "Couldn't save before leaving",
    code: "leave_save_failed",
    fallbackDetail: "Saving the prompt failed before leaving.",
  });
}

export function renameSessionError(err: unknown): PromptEditorSessionError {
  return sessionErrorFromUnknown("save", err, {
    title: "Couldn't rename",
    code: "rename_failed",
    fallbackDetail: "Saving the document title failed.",
  });
}

export const HYDRATE_FALLBACK_WARNING: PromptEditorSessionError = {
  phase: "hydrate",
  title: "Showing plain text",
  detail:
    "We couldn't restore the rich layout; showing plain text. Your words are still here.",
  code: "hydrate_fallback",
};
