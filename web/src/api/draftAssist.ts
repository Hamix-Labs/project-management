import type {
  DraftAssistCancelRunResult,
  DraftAssistReady,
  DraftAssistSession,
  DraftAssistSnapshot,
  DraftAssistSnapshotUpdate,
  DraftAssistStartRunResult,
} from "@/types/draftAssist";
import {
  parseDraftAssistCancelRunResult,
  parseDraftAssistReady,
  parseDraftAssistSession,
  parseDraftAssistSnapshotUpdate,
  parseDraftAssistStartRunResult,
} from "./parseTaskApiDraftAssist";
import { apiErrorFromResponse, fetchWithTimeout, jsonHeaders } from "./shared";

/**
 * All `/draft-assist/*` HTTP + SSE transport lives here (ADR aligns
 * with `web/src/api/` fetch ownership; see `cmd/checkstandards`).
 */

const draftAssistBasePath = "/draft-assist";

/** GET /draft-assist/ready — SPA cheap probe; does not throw for 4xx (surfaces reason). */
export async function readyProbe(options?: { signal?: AbortSignal }): Promise<DraftAssistReady> {
  const res = await fetchWithTimeout(`${draftAssistBasePath}/ready`, {
    headers: { Accept: "application/json" },
    signal: options?.signal,
  });
  if (!res.ok) throw await apiErrorFromResponse(res);
  return parseDraftAssistReady((await res.json()) as unknown);
}

/** POST /draft-assist/sessions — lazy on first `open(snapshot)`. */
export async function createSession(
  input: { worktree_id?: string; snapshot: DraftAssistSnapshot },
  options?: { signal?: AbortSignal },
): Promise<DraftAssistSession> {
  const res = await fetchWithTimeout(`${draftAssistBasePath}/sessions`, {
    method: "POST",
    headers: jsonHeaders,
    body: JSON.stringify({
      worktree_id: input.worktree_id ?? "",
      snapshot: input.snapshot,
    }),
    signal: options?.signal,
  });
  if (!res.ok) throw await apiErrorFromResponse(res);
  return parseDraftAssistSession((await res.json()) as unknown);
}

/** GET /draft-assist/sessions/{id} — reconcile server state (rarely needed once SSE is up). */
export async function getSession(
  sessionId: string,
  options?: { signal?: AbortSignal },
): Promise<DraftAssistSession> {
  const id = assertSessionId(sessionId);
  const res = await fetchWithTimeout(
    `${draftAssistBasePath}/sessions/${encodeURIComponent(id)}`,
    {
      headers: { Accept: "application/json" },
      signal: options?.signal,
    },
  );
  if (!res.ok) throw await apiErrorFromResponse(res);
  return parseDraftAssistSession((await res.json()) as unknown);
}

/** PUT /draft-assist/sessions/{id}/snapshot — replace snapshot on operator edits. */
export async function updateSnapshot(
  sessionId: string,
  snapshot: DraftAssistSnapshot,
  options?: { signal?: AbortSignal },
): Promise<DraftAssistSnapshotUpdate> {
  const id = assertSessionId(sessionId);
  const res = await fetchWithTimeout(
    `${draftAssistBasePath}/sessions/${encodeURIComponent(id)}/snapshot`,
    {
      method: "PUT",
      headers: jsonHeaders,
      body: JSON.stringify(snapshot),
      signal: options?.signal,
    },
  );
  if (!res.ok) throw await apiErrorFromResponse(res);
  return parseDraftAssistSnapshotUpdate((await res.json()) as unknown);
}

/**
 * POST /draft-assist/sessions/{id}/runs — returns 202 immediately with
 * a run id. A concurrent run yields HTTP 409 which surfaces as
 * {@link ApiError} with `.status === 409` so callers can render
 * "already running" without regex-matching the message.
 */
export async function startRun(
  sessionId: string,
  input: { user_message: string; snapshot?: DraftAssistSnapshot },
  options?: { signal?: AbortSignal },
): Promise<DraftAssistStartRunResult> {
  const id = assertSessionId(sessionId);
  const body: Record<string, unknown> = { user_message: input.user_message };
  if (input.snapshot !== undefined) {
    body.snapshot = input.snapshot;
  }
  const res = await fetchWithTimeout(
    `${draftAssistBasePath}/sessions/${encodeURIComponent(id)}/runs`,
    {
      method: "POST",
      headers: jsonHeaders,
      body: JSON.stringify(body),
      signal: options?.signal,
    },
  );
  if (!res.ok) throw await apiErrorFromResponse(res);
  return parseDraftAssistStartRunResult((await res.json()) as unknown);
}

/** POST /draft-assist/sessions/{id}/runs/{runId}/cancel — SSE emits cancelling then done{cancelled}. */
export async function cancelRun(
  sessionId: string,
  runId: string,
  options?: { signal?: AbortSignal },
): Promise<DraftAssistCancelRunResult> {
  const sid = assertSessionId(sessionId);
  const rid = assertRunId(runId);
  const res = await fetchWithTimeout(
    `${draftAssistBasePath}/sessions/${encodeURIComponent(sid)}/runs/${encodeURIComponent(rid)}/cancel`,
    {
      method: "POST",
      headers: jsonHeaders,
      body: "{}",
      signal: options?.signal,
    },
  );
  if (!res.ok) throw await apiErrorFromResponse(res);
  return parseDraftAssistCancelRunResult((await res.json()) as unknown);
}

/**
 * DELETE /draft-assist/sessions/{id} — best-effort teardown. Callers
 * fire-and-forget on modal close; 404 is tolerated because the server
 * may have already GC-ed an idle session.
 */
export async function deleteSession(
  sessionId: string,
  options?: { signal?: AbortSignal },
): Promise<void> {
  const id = assertSessionId(sessionId);
  const res = await fetchWithTimeout(
    `${draftAssistBasePath}/sessions/${encodeURIComponent(id)}`,
    {
      method: "DELETE",
      headers: { Accept: "application/json" },
      signal: options?.signal,
    },
  );
  if (res.status === 404 || res.status === 204) return;
  if (!res.ok) throw await apiErrorFromResponse(res);
}

/**
 * Build the SSE URL for a session. Kept as a helper so tests can assert
 * the URL shape without opening a real EventSource.
 */
export function draftAssistEventsUrl(sessionId: string): string {
  const id = assertSessionId(sessionId);
  return `${draftAssistBasePath}/sessions/${encodeURIComponent(id)}/events`;
}

/**
 * SSE transport factory for GET /draft-assist/sessions/{id}/events.
 * All EventSource construction lives under `web/src/api/` (parallels
 * `openTaskEventsSource`; enforced by `cmd/checkstandards`).
 */
export function openDraftAssistEventsSource(sessionId: string): EventSource {
  return new EventSource(draftAssistEventsUrl(sessionId));
}

function assertSessionId(id: string): string {
  const t = id.trim();
  if (t === "") throw new Error("draft-assist session id is required");
  return t;
}

function assertRunId(id: string): string {
  const t = id.trim();
  if (t === "") throw new Error("draft-assist run id is required");
  return t;
}
