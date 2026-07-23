/** Mirrors pkgs/gitinventory/store.TaskBranchName — hamix/task-<first 8 hex>. */
export function taskBranchName(taskId: string): string {
  let hex = taskId.trim().replace(/-/g, "").toLowerCase();
  if (hex.length > 8) {
    hex = hex.slice(0, 8);
  }
  if (hex === "") {
    hex = "00000000";
  }
  return `hamix/task-${hex}`;
}

/** Mirrors pkgs/gitinventory.BranchPathSlug for managed worktree folder names. */
export function branchPathSlug(branch: string): string {
  const trimmed = branch.trim();
  if (trimmed === "") {
    return "branch";
  }
  let out = "";
  let prevDash = false;
  for (const ch of trimmed) {
    const code = ch.codePointAt(0) ?? 0;
    if (ch === "/" || ch === "\\") {
      if (!prevDash) {
        out += "-";
        prevDash = true;
      }
      continue;
    }
    const isAlnum =
      (code >= 48 && code <= 57) ||
      (code >= 65 && code <= 90) ||
      (code >= 97 && code <= 122);
    if (isAlnum || ch === "." || ch === "_" || ch === "-") {
      out += ch;
      prevDash = ch === "-";
      continue;
    }
    if (!prevDash) {
      out += "-";
      prevDash = true;
    }
  }
  out = out.replace(/^-+|-+$/g, "");
  return out === "" ? "branch" : out;
}

/** Predicted worktree display name before allocate completes. */
export function predictedTaskWorktreeName(taskId: string): string {
  return branchPathSlug(taskBranchName(taskId));
}
