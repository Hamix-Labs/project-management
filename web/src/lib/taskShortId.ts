/** Short display form of a task UUID (first 8 hex chars). */
export function shortId(id: string): string {
  return id.length > 8 ? id.slice(0, 8) : id;
}

/**
 * Canonical short reference for a task in list rows, board cards,
 * depends-on chips, and timeline highlights. Prefers the server-assigned
 * per-project sequential number as `#N` (see docs/data-model.md); falls
 * back to the shortened UUID for legacy rows or global tasks that were
 * created before the backfill migration ran.
 *
 * `number` accepts `null`/`undefined` to keep call sites tolerant of
 * partial task shapes coming off SSE hints and activity envelopes.
 */
export function taskDisplayRef(task: {
  id: string;
  number?: number | null;
}): string {
  if (typeof task.number === "number" && Number.isFinite(task.number)) {
    return `#${task.number}`;
  }
  return shortId(task.id);
}
