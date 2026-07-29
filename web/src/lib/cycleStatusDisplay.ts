import type { StatusMeta, StatusTone } from "@/lib/taskStatusDisplay";
import type { CycleStatus } from "@/types/cycle";

type CycleStatusMeta = Omit<StatusMeta, "order"> & { tone: StatusTone };

/**
 * Presentation metadata for cycle outcome badges — same tone vocabulary as
 * task STATUS_META so Running/Succeeded/Failed match StatusBadge chrome.
 */
export const CYCLE_STATUS_META: Record<CycleStatus, CycleStatusMeta> = {
  running: { label: "Running", tone: "info", pulse: true },
  succeeded: { label: "Succeeded", tone: "success" },
  failed: { label: "Failed", tone: "danger" },
  aborted: { label: "Aborted", tone: "neutral" },
};
