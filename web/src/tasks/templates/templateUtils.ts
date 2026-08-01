import { maxTemplateInstantiateCountPerItem } from "@/api";
import { formatRelativeTime } from "@/shared/time/relativeTime";
import { isRowActionExcluded } from "@/tasks/components/saved-entities/rowActionUtils";

type InstantiateTemplatesBatchResult = {
  accepted: boolean;
  total: number;
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
  if (!result.accepted && result.errors.length > 0) {
    return result.errors.map((entry) => `${entry.template_id}: ${entry.error}`).join(" ");
  }
  if (result.errors.length > 0) {
    return `Accepted ${result.total} task(s). Skipped: ${result.errors
      .map((entry) => entry.template_id)
      .join(", ")}`;
  }
  if (!result.accepted || result.total < 1) {
    return "Could not accept template create request.";
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
