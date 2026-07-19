/** Mirrors pkgs/projects/domain MaxProjectContext* limits. */
export const MAX_PROJECT_CONTEXT_BODY_BYTES = 512 * 1024;
export const MAX_PROJECT_CONTEXT_TITLE_CHARS = 200;

export const MEMORY_IMPORT_ACCEPT =
  ".txt,.md,text/plain,text/markdown" as const;

export const MEMORY_IMPORT_EXTENSIONS = [".txt", ".md"] as const;
