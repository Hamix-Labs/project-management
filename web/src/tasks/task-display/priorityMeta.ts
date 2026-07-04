import type { Priority } from "@/types";
import { priorityListLabel } from "./priorityListLabel";

export type PriorityTone = "neutral" | "info" | "warning" | "danger";

export type PriorityMeta = {
  label: string;
  tone: PriorityTone;
  weight: number;
};

/** Presentation metadata for priority badges and client-side list sort. */
export const PRIORITY_META: Record<Priority, PriorityMeta> = {
  low: { label: priorityListLabel("low"), tone: "neutral", weight: 1 },
  medium: { label: priorityListLabel("medium"), tone: "info", weight: 2 },
  high: { label: priorityListLabel("high"), tone: "warning", weight: 3 },
  critical: { label: priorityListLabel("critical"), tone: "danger", weight: 4 },
};
