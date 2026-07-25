/**
 * advancedSummaryLine builds the one-line caption shown next to the
 * collapsed "More options" disclosure on the new-task modal. It lets the
 * operator see the effective values for the secondary fields (agent,
 * schedule, tags, milestone, dependencies) without expanding the panel.
 *
 * Returns a stable fallback when every input is at its default — that
 * copy doubles as the affordance description so the disclosure never
 * reads as empty chrome.
 */
export function advancedSummaryLine(input: {
  runner: string;
  cursorModel: string;
  schedule: string | null;
  tagsCsv: string;
  milestone: string;
  dependsOn: string[];
  /** When false, schedule is omitted from the collapsed summary (launch gate). */
  includeSchedule?: boolean;
  /** When false, tags are omitted from the collapsed summary (launch gate). */
  includeTags?: boolean;
  /** When false, milestone/deps are omitted from the collapsed summary (launch gate). */
  includeDependencies?: boolean;
}): string {
  const parts: string[] = [];
  const includeSchedule = input.includeSchedule ?? true;
  const includeTags = input.includeTags ?? true;
  const includeDependencies = input.includeDependencies ?? true;

  const runnerLabel = runnerDisplayLabel(input.runner);
  const modelLabel = input.cursorModel.trim();
  if (modelLabel) {
    parts.push(`${runnerLabel} · ${modelLabel}`);
  }

  if (includeSchedule && input.schedule) {
    parts.push("Scheduled");
  }

  if (includeTags) {
    const tagCount = countCsv(input.tagsCsv);
    if (tagCount > 0) {
      parts.push(`${tagCount} ${tagCount === 1 ? "tag" : "tags"}`);
    }
  }

  if (includeDependencies) {
    if (input.milestone.trim()) {
      parts.push("Milestone");
    }

    const depCount = input.dependsOn.length;
    if (depCount > 0) {
      parts.push(`${depCount} ${depCount === 1 ? "dep" : "deps"}`);
    }
  }

  if (parts.length === 0) {
    return advancedSummaryFallback(includeSchedule, includeTags, includeDependencies);
  }

  return parts.join(" · ");
}

function advancedSummaryFallback(
  includeSchedule: boolean,
  includeTags: boolean,
  includeDependencies: boolean,
): string {
  const labels = ["Agent"];
  if (includeSchedule) labels.push("schedule");
  if (includeTags) labels.push("tags");
  if (includeDependencies) labels.push("dependencies");
  return labels.join(", ");
}

function runnerDisplayLabel(id: string): string {
  if (id === "cursor") return "Cursor CLI";
  return id || "Cursor CLI";
}

function countCsv(csv: string): number {
  if (!csv) return 0;
  return csv
    .split(",")
    .map((s) => s.trim())
    .filter(Boolean).length;
}
