/** Mirrors pkgs/projects/domain MaxProjectContext* limits. */
export const MAX_PROJECT_CONTEXT_BODY_BYTES = 512 * 1024;
export const MAX_PROJECT_CONTEXT_TITLE_CHARS = 200;
export const MAX_PROJECT_CONTEXT_DESCRIPTION_CHARS = 400;
export const MAX_PROJECT_CONTEXT_TAG_CHARS = 40;

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

/** Returns an error message when tag is empty or over the char limit. */
export function validateProjectContextTag(tag: string): string | null {
  const trimmed = tag.trim();
  if (!trimmed) {
    return "Tag is required.";
  }
  if ([...trimmed].length > MAX_PROJECT_CONTEXT_TAG_CHARS) {
    return `Tag must be ${MAX_PROJECT_CONTEXT_TAG_CHARS} characters or fewer.`;
  }
  return null;
}
