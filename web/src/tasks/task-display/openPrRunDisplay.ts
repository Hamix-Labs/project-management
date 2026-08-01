/** Display label when a running cycle is an approve-and-open-PR run. */
export const CREATING_PR_STATUS_LABEL = "Creating PR";

/** True when cycle meta stamps run_kind=open_pr. */
export function isOpenPrRunKind(meta: Record<string, unknown> | null | undefined): boolean {
  return typeof meta?.run_kind === "string" && meta.run_kind === "open_pr";
}
