import { maxTemplateInstantiateCountPerItem } from "@/api";
import { formatRelativeTime } from "@/shared/time/relativeTime";
import { isRowActionExcluded } from "@/tasks/components/saved-entities/rowActionUtils";
import type { Task } from "@/types";

type InstantiateTemplatesBatchResult = {
  tasks: Task[];
  errors: { template_id: string; error: string }[];
};

export function clampInstanceCount(value: number): number {
  if (!Number.isFinite(value)) return 1;
  return Math.min(
    maxTemplateInstantiateCountPerItem,
    Math.max(1, Math.floor(value)),
  );
}

export function sumSelectedInstanceCounts(
  selectedIds: string[],
  instanceCounts: Record<string, number>,
): number {
  return selectedIds.reduce(
    (sum, id) => sum + (instanceCounts[id] ?? 1),
    0,
  );
}

export function formatInstantiateBatchError(
  result: InstantiateTemplatesBatchResult,
): string | null {
  if (result.errors.length > 0 && result.tasks.length === 0) {
    return result.errors.map((entry) => `${entry.template_id}: ${entry.error}`).join(" ");
  }
  if (result.errors.length > 0) {
    return `Created ${result.tasks.length} task(s). Failed: ${result.errors
      .map((entry) => entry.template_id)
      .join(", ")}`;
  }
  return null;
}

/** @deprecated Use formatRelativeTime from @/shared/time/relativeTime */
export function formatTemplateRelativeTime(
  iso: string | null | undefined,
  now: Date = new Date(),
): string {
  return formatRelativeTime(iso ?? "", now);
}

export function isTemplateRowActionExcluded(target: EventTarget | null): boolean {
  if (!(target instanceof Element)) return true;
  return isRowActionExcluded(target);
}

/** Last path segment for repo labels (e.g. `C:/proj/hamix` → `hamix`). */
export function repositoryBasename(path: string): string {
  const normalized = path.trim().replace(/\\/g, "/").replace(/\/+$/, "");
  if (normalized === "") return path;
  const slash = normalized.lastIndexOf("/");
  return slash >= 0 ? normalized.slice(slash + 1) : normalized;
}

export function shortIdLabel(id: string): string {
  const trimmed = id.trim();
  if (trimmed.length <= 8) return trimmed;
  return trimmed.slice(0, 8);
}

/** Resolve a display label from a name map, falling back to a short id. */
export function templateBindingLabel(
  id: string | undefined,
  nameById: Map<string, string>,
): string | null {
  if (!id?.trim()) return null;
  return nameById.get(id) ?? shortIdLabel(id);
}
