import type { Status } from "@/types";

/** Display label when a running cycle is an approve-and-open-PR run. */
export const CREATING_PR_STATUS_LABEL = "Creating PR";

/** True when cycle meta stamps run_kind=open_pr. */
export function isOpenPrRunKind(
  meta: Record<string, unknown> | null | undefined,
): boolean {
  return typeof meta?.run_kind === "string" && meta.run_kind === "open_pr";
}

/** Whether the toolbar should show Creating PR instead of Running/Ready. */
export function shouldShowCreatingPrLabel(input: {
  mutationPending: boolean;
  sessionActive: boolean;
  hasRunningOpenPrCycle: boolean;
}): boolean {
  return (
    input.mutationPending ||
    input.sessionActive ||
    input.hasRunningOpenPrCycle
  );
}

/**
 * True when task status means the open-PR sticky session should clear
 * (no longer queued or running the open-PR visit).
 */
export function openPrSessionClearedByStatus(status: Status): boolean {
  return status !== "ready" && status !== "running";
}
