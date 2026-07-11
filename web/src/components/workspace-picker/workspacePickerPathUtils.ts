import type { WorkspaceBrowseRoot } from "@/api/settingsBrowse";

export type Crumb = { label: string; path: string };

const USER_FOLDER_CATEGORIES = new Set([
  "documents",
  "desktop",
  "downloads",
  "pictures",
  "music",
  "videos",
]);

export function normalizeBrowsePath(path: string): string {
  return path.trim().replace(/[\\/]+$/, "");
}

export function isBrowseRootPath(roots: WorkspaceBrowseRoot[], path: string): boolean {
  const normalized = normalizeBrowsePath(path).toLowerCase();
  return roots.some(
    (root) => normalizeBrowsePath(root.path).toLowerCase() === normalized,
  );
}

export function partitionBrowseRoots(roots: WorkspaceBrowseRoot[]): {
  workspace: WorkspaceBrowseRoot[];
  userFolders: WorkspaceBrowseRoot[];
} {
  const workspace: WorkspaceBrowseRoot[] = [];
  const userFolders: WorkspaceBrowseRoot[] = [];
  for (const root of roots) {
    if (root.category && USER_FOLDER_CATEGORIES.has(root.category)) {
      userFolders.push(root);
    } else {
      workspace.push(root);
    }
  }
  return { workspace, userFolders };
}

// Splits the current path into breadcrumb segments anchored to whichever
// starting location contains it. Without anchoring to a root, deep paths
// like /Users/x/code/proj would surface every system folder as a crumb,
// which is noise — what matters is "Home > code > proj".
export function computeCrumbs(
  roots: WorkspaceBrowseRoot[],
  currentBrowsePath: string,
): Crumb[] {
  const trimmed = currentBrowsePath.trim();
  if (trimmed === "") return [];
  const root = roots.find(
    (r) =>
      trimmed === r.path ||
      trimmed.startsWith(`${r.path}/`) ||
      trimmed.startsWith(`${r.path}\\`),
  );
  if (!root) {
    return [{ label: trimmed, path: trimmed }];
  }
  const crumbs: Crumb[] = [{ label: root.label, path: root.path }];
  if (trimmed === root.path) return crumbs;
  const sep = trimmed.includes("\\") ? "\\" : "/";
  const rel = trimmed.slice(root.path.length).replace(/^[\\/]+/, "");
  const parts = rel.split(/[\\/]+/).filter(Boolean);
  let acc = root.path;
  for (const part of parts) {
    acc = `${acc}${sep}${part}`;
    crumbs.push({ label: part, path: acc });
  }
  return crumbs;
}
