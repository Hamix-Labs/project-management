import { useEffect, useState } from "react";
import { maxTemplateInstantiateCountPerItem } from "@/api";
import type { Task } from "@/types";

type InstantiateTemplatesBatchResult = {
  tasks: Task[];
  errors: { template_id: string; error: string }[];
};

export function useDebouncedTrimmedValue(value: string, delayMs: number): string {
  const [debounced, setDebounced] = useState(value.trim());

  useEffect(() => {
    const timer = window.setTimeout(() => setDebounced(value.trim()), delayMs);
    return () => window.clearTimeout(timer);
  }, [value, delayMs]);

  return debounced;
}

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

/** Compact relative time for template rows: "2d ago", not "2 d ago". */
export function formatTemplateRelativeTime(
  iso: string | null | undefined,
  now: Date = new Date(),
): string {
  if (!iso) return "";
  const then = new Date(iso);
  const t = then.getTime();
  if (!Number.isFinite(t)) return "";

  const deltaMs = now.getTime() - t;
  if (deltaMs < 45_000) return "just now";

  const minutes = Math.floor(deltaMs / 60_000);
  if (minutes < 60) return `${minutes}m ago`;

  const hours = Math.floor(deltaMs / 3_600_000);
  if (hours < 24) return `${hours}h ago`;

  const days = Math.floor(deltaMs / 86_400_000);
  if (days < 7) return `${days}d ago`;

  const weeks = Math.floor(days / 7);
  if (days < 30) return `${weeks}w ago`;

  const months = Math.floor(days / 30);
  if (days < 365) return `${months}mo ago`;

  const years = Math.floor(days / 365);
  return `${years}y ago`;
}

export function isTemplateRowActionExcluded(target: EventTarget | null): boolean {
  if (!(target instanceof Element)) return true;
  return Boolean(target.closest("button, input, label, select, a"));
}
