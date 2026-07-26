import type { ProjectContextItem } from "@/types";

export type ProjectContextTagGroup = {
  label: string;
  items: ProjectContextItem[];
};

/** Unique tags from items, preserving first-seen casing, sorted A–Z. */
export function collectProjectContextTags(
  items: ProjectContextItem[],
): string[] {
  const seen = new Map<string, string>();
  for (const item of items) {
    const display = item.tag.trim();
    const key = display.toLowerCase();
    if (!key || seen.has(key)) continue;
    seen.set(key, display);
  }
  return [...seen.values()].sort((a, b) => a.localeCompare(b));
}

/** Group items by tag (case-insensitive); empty tag → General. */
export function groupProjectContextByTag(
  items: ProjectContextItem[],
): ProjectContextTagGroup[] {
  const map = new Map<string, ProjectContextTagGroup>();
  for (const item of items) {
    const display = item.tag.trim() || "General";
    const key = display.toLowerCase();
    const existing = map.get(key);
    if (existing) {
      existing.items.push(item);
    } else {
      map.set(key, { label: display, items: [item] });
    }
  }
  return [...map.values()].sort((a, b) => a.label.localeCompare(b.label));
}
