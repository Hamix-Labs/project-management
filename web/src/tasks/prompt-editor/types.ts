/** Identity for a prompt document opened in the full-page Prompt Editor. */
export type PromptSourceKind = "draft" | "task" | "template" | "ephemeral";

export type PromptDocumentRef = {
  kind: PromptSourceKind;
  id: string;
};

/** Optional context for @-mentions and return navigation. */
export type PromptEditorLaunchContext = {
  worktreeId?: string;
  returnPath?: string;
  /** When true, Done resumes the suspended compose modal. */
  resumeCompose?: boolean;
  /** Polish dialog should reopen after Done. */
  resumePolish?: boolean;
  polishTaskId?: string;
  placeholder?: string;
  title?: string;
  /** Prefer this HTML on first paint (compose form may be ahead of server). */
  seedHtml?: string;
};

export type PromptDocumentSnapshot = {
  html: string;
  /** Display name when known (draft/template). */
  name?: string;
  worktreeId?: string;
};

export type PromptDocumentAdapter = {
  load: (signal?: AbortSignal) => Promise<PromptDocumentSnapshot>;
  save: (html: string, signal?: AbortSignal) => Promise<void>;
};

export const PROMPT_EDITOR_LAUNCH_KEY = "hamix:prompt-editor-launch";
export const PROMPT_EPHEMERAL_PREFIX = "hamix:prompt-ephemeral:";
export const PROMPT_RETURN_KEY = "hamix:prompt-editor-return";

export type PromptEditorReturnPayload = {
  resumeCompose?: boolean;
  resumePolish?: boolean;
  polishTaskId?: string;
  returnPath: string;
  /** Latest HTML written on Done (compose sync). */
  html?: string;
};
