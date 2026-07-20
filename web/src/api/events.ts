/**
 * SSE transport factory for GET /events.
 * Connection policy and frame handling stay in the tasks vertical;
 * this module owns constructing EventSource so network transports
 * remain under web/src/api/ (same ownership as fetch).
 */
export function openTaskEventsSource(url = "/events"): EventSource {
  return new EventSource(url);
}
