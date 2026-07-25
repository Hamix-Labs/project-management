export function formatBulkFailure(failedCount: number, attempted: number): string {
  return `${failedCount} of ${attempted} reschedules failed. The successful ones already updated; the failed rows kept their previous schedule. Try again or check the task detail pages for details.`;
}

/**
 * Renamed from `formatBulkDeleteFailure` when the retired hard-delete
 * flow was replaced by the reversible `POST /tasks/{id}/close`
 * lifecycle (see docs/api.md). Copy no longer implies data loss.
 */
export function formatBulkCloseFailure(failedCount: number, attempted: number): string {
  return `${failedCount} of ${attempted} closes failed. Tasks that were closed stay closed (you can reopen them); try again for the rest.`;
}
