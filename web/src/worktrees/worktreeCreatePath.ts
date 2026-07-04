/** Joins a parent directory and new folder name into an absolute worktree path. */
export function joinWorktreeCreatePath(parentPath: string, folderName: string): string | null {
  const parent = parentPath.trim().replace(/[\\/]+$/, "");
  const folder = folderName.trim().replace(/^[\\/]+/, "").replace(/[\\/]+$/, "");
  if (parent === "" || folder === "" || folder.includes("/") || folder.includes("\\")) {
    return null;
  }
  const sep = parent.includes("\\") ? "\\" : "/";
  return `${parent}${sep}${folder}`;
}
