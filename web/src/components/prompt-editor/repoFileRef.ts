export type RepoFileRef = {
  path: string;
  lineStart?: number;
  lineEnd?: number;
};

export function parseOptionalLine(raw: string | null | undefined): number | undefined {
  if (raw == null || raw === "") return undefined;
  const n = parseInt(raw, 10);
  return Number.isFinite(n) && n > 0 ? n : undefined;
}

export function normalizeRepoPath(path: string): string {
  return path.replace(/\\/g, "/").trim();
}

export function splitRepoPath(path: string): { fileName: string; dirPath: string } {
  const norm = normalizeRepoPath(path);
  const i = norm.lastIndexOf("/");
  if (i < 0) return { fileName: norm, dirPath: "" };
  return { fileName: norm.slice(i + 1), dirPath: norm.slice(0, i + 1) };
}

export function formatLineRangeLabel(
  lineStart?: number,
  lineEnd?: number,
): string | null {
  if (lineStart == null) return null;
  if (lineEnd == null || lineEnd === lineStart) return `line ${lineStart}`;
  return `lines ${lineStart}–${lineEnd}`;
}

export function sliceFileLines(
  content: string,
  lineStart?: number,
  lineEnd?: number,
): { start: number; lines: string[] } {
  const all = content.split("\n");
  if (lineStart == null) {
    return { start: 1, lines: all.slice(0, 40) };
  }
  const start = Math.max(1, lineStart);
  const end = Math.max(start, lineEnd ?? lineStart);
  return {
    start,
    lines: all.slice(start - 1, end),
  };
}
