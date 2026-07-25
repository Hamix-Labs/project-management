/** Short display form of a task UUID (first 8 hex chars). */
export function shortId(id: string): string {
  return id.length > 8 ? id.slice(0, 8) : id;
}
