import type { Task } from "@/types";

export { shortId } from "@/lib/taskShortId";

export const MAX_TYPEAHEAD_RESULTS = 8;
export const TYPEAHEAD_BLUR_DELAY_MS = 120;

export function filterTypeaheadCandidates(
  projectTasks: Task[],
  query: string,
  selectedSet: Set<string>,
  maxResults: number,
): Task[] {
  const q = query.trim().toLowerCase();
  const candidates = projectTasks.filter((t) => !selectedSet.has(t.id));
  if (!q) return candidates.slice(0, maxResults);
  const hits: Task[] = [];
  for (const t of candidates) {
    if (
      t.title.toLowerCase().includes(q) ||
      t.id.toLowerCase().startsWith(q)
    ) {
      hits.push(t);
      if (hits.length >= maxResults) break;
    }
  }
  return hits;
}

export function filterBrowseCandidates(
  projectTasks: Task[],
  browseQuery: string,
): Task[] {
  const q = browseQuery.trim().toLowerCase();
  if (!q) return projectTasks;
  return projectTasks.filter(
    (t) =>
      t.title.toLowerCase().includes(q) || t.id.toLowerCase().includes(q),
  );
}

export function buildDependsOnHelperCopy(
  hasProject: boolean,
  isLoading: boolean,
  projectTaskCount: number,
): string {
  if (!hasProject) return "Pick a project first to add dependencies.";
  if (isLoading) return "Loading project tasks…";
  if (projectTaskCount === 0) return "No tasks exist in this project yet.";
  return "Other tasks that must complete before the agent picks this one up.";
}
