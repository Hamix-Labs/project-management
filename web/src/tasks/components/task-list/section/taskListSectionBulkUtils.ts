export function formatBulkFailure(failedCount: number, attempted: number): string {
  return `${failedCount} of ${attempted} reschedules failed. The successful ones already updated; the failed rows kept their previous schedule. Try again or check the task detail pages for details.`;
}

export function formatBulkDeleteFailure(failedCount: number, attempted: number): string {
  return `${failedCount} of ${attempted} deletes failed. Tasks that were removed stay deleted; try again for the rest.`;
}
