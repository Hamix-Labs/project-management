/** Shared label for @repo file chips (TipTap + BlockNote). */
export function repoFileMentionLabel(attrs: {
  path: string;
  lineStart?: number | null;
  lineEnd?: number | null;
}): string {
  const { path, lineStart, lineEnd } = attrs;
  if (
    lineStart != null &&
    lineEnd != null &&
    Number.isFinite(lineStart) &&
    Number.isFinite(lineEnd)
  ) {
    return `@${path}(${lineStart}-${lineEnd})`;
  }
  return `@${path}`;
}
