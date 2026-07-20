/** Mirrors pkgs/projects/domain MaxProjectContext* limits. */
export const MAX_PROJECT_CONTEXT_BODY_BYTES = 512 * 1024;
export const MAX_PROJECT_CONTEXT_TITLE_CHARS = 200;
export const MAX_PROJECT_CONTEXT_DESCRIPTION_CHARS = 400;

export const MEMORY_IMPORT_ACCEPT =
  ".txt,.md,text/plain,text/markdown" as const;

export const MEMORY_IMPORT_EXTENSIONS = [".txt", ".md"] as const;

/** Returns an error message when description exceeds the char limit; empty is allowed. */
export function validateProjectContextDescription(
  description: string,
): string | null {
  const trimmed = description.trim();
  if ([...trimmed].length > MAX_PROJECT_CONTEXT_DESCRIPTION_CHARS) {
    return `Description must be ${MAX_PROJECT_CONTEXT_DESCRIPTION_CHARS} characters or fewer.`;
  }
  return null;
}
