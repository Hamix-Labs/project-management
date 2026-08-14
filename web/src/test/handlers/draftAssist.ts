import { http, HttpResponse } from "msw";
import {
  DRAFT_ASSIST_SCHEMA_VERSION,
  type DraftAssistSnapshot,
} from "@/types/draftAssist";

/**
 * MSW handlers for `/draft-assist/*`. Covers:
 *  - happy path: ready → create → snapshot → start run → cancel → delete
 *  - 409 (concurrent run)
 *  - missing runner ready probe
 *  - cancel two-frame body preview (status=cancelling then done{cancelled})
 *  - replay session detail for reconnect assertions
 *
 * SSE frames themselves are not delivered by MSW's fetch interceptor
 * (browser EventSource bypasses MSW under Vitest/jsdom); hook tests
 * stub `EventSource` via `web/src/test/browserMocks.ts` and drive
 * frames directly. These handlers only cover the JSON side of the API.
 */

const draftAssistBase = "/draft-assist";

export type DraftAssistSessionFixture = {
  id: string;
  nonce: string;
  worktree_id?: string;
  snapshot: DraftAssistSnapshot;
  created_at: string;
  updated_at?: string;
};

export function makeDraftAssistSessionFixture(
  overrides: Partial<DraftAssistSessionFixture> = {},
): DraftAssistSessionFixture {
  return {
    id: "da-sess-1",
    nonce: "nonce-1",
    worktree_id: "wt-1",
    snapshot: { title: "", prompt: "" },
    created_at: "2026-08-14T00:00:00Z",
    ...overrides,
  };
}

/** GET /draft-assist/ready — ready with a runner name. */
export function draftAssistReadyOk(runner: "fake" | "sdk" = "fake") {
  return http.get(`${draftAssistBase}/ready`, () =>
    HttpResponse.json({ ready: true, runner }),
  );
}

/** GET /draft-assist/ready — no runner configured (missing_key / no_runner / sidecar_down). */
export function draftAssistReadyMissing(
  reason: "no_runner" | "missing_key" | "sidecar_down" = "no_runner",
  runner: "missing" | "sdk" = "missing",
) {
  return http.get(`${draftAssistBase}/ready`, () =>
    HttpResponse.json({ ready: false, runner, reason }),
  );
}

/** POST /draft-assist/sessions — 201 create. */
export function draftAssistCreateSession(
  fixture: DraftAssistSessionFixture = makeDraftAssistSessionFixture(),
) {
  return http.post(`${draftAssistBase}/sessions`, () =>
    HttpResponse.json(fixture, { status: 201 }),
  );
}

/** POST /draft-assist/sessions — capture request body and return fixture. */
export function draftAssistCreateSessionCapture(
  onPost: (body: string) => void,
  fixture: DraftAssistSessionFixture = makeDraftAssistSessionFixture(),
) {
  return http.post(`${draftAssistBase}/sessions`, async ({ request }) => {
    onPost(await request.text());
    return HttpResponse.json(fixture, { status: 201 });
  });
}

/** GET /draft-assist/sessions/{id} — return current server-side snapshot. */
export function draftAssistGetSession(
  fixture: DraftAssistSessionFixture = makeDraftAssistSessionFixture(),
) {
  return http.get(`${draftAssistBase}/sessions/${fixture.id}`, () =>
    HttpResponse.json({
      ...fixture,
      updated_at: fixture.updated_at ?? fixture.created_at,
    }),
  );
}

/** PUT /draft-assist/sessions/{id}/snapshot — 200 replace. */
export function draftAssistUpdateSnapshot(
  sessionId: string,
  snapshot: DraftAssistSnapshot,
) {
  return http.put(`${draftAssistBase}/sessions/${sessionId}/snapshot`, () =>
    HttpResponse.json({ id: sessionId, snapshot }),
  );
}

/** POST /draft-assist/sessions/{id}/runs — 202 { run_id }. */
export function draftAssistStartRunOk(sessionId: string, runId = "run-1") {
  return http.post(`${draftAssistBase}/sessions/${sessionId}/runs`, () =>
    HttpResponse.json({ run_id: runId }, { status: 202 }),
  );
}

/** POST /draft-assist/sessions/{id}/runs — 409 concurrent run. */
export function draftAssistStartRunConflict(sessionId: string) {
  return http.post(`${draftAssistBase}/sessions/${sessionId}/runs`, () =>
    HttpResponse.json({ error: "run already active" }, { status: 409 }),
  );
}

/** POST /draft-assist/sessions/{id}/runs/{runId}/cancel — 202 cancelling. */
export function draftAssistCancelRunOk(sessionId: string, runId = "run-1") {
  return http.post(
    `${draftAssistBase}/sessions/${sessionId}/runs/${runId}/cancel`,
    () =>
      HttpResponse.json(
        { run_id: runId, status: "cancelling" },
        { status: 202 },
      ),
  );
}

/**
 * The SSE cancel path emits two frames on the wire (`status=cancelling`,
 * then `done{status=cancelled}`). MSW does not drive SSE for these
 * hook tests; export the raw frame preview so the stream-hook suite
 * can feed identical bytes to its mock EventSource.
 */
export const draftAssistCancelFrames = [
  {
    id: 10,
    kind: "status" as const,
    run_id: "run-1",
    at: "2026-08-14T00:00:05Z",
    data: { status: "cancelling" as const },
  },
  {
    id: 11,
    kind: "done" as const,
    run_id: "run-1",
    at: "2026-08-14T00:00:06Z",
    data: { status: "cancelled" as const },
  },
];

/**
 * Ring-buffer replay preview for reconnect tests. First array is the
 * initial run of frames the server would send on the fresh connection;
 * the second array is what the server would replay after
 * `Last-Event-ID` reconnect (overlapping ids should be deduplicated
 * by the hook).
 */
export function draftAssistReplayFrames(sessionId: string) {
  return {
    initial: [
      draftAssistSessionFrame(sessionId, 1),
      { id: 2, kind: "status" as const, run_id: "run-1", at: "2026-08-14T00:00:01Z", data: { status: "thinking" as const } },
      { id: 3, kind: "token" as const, run_id: "run-1", at: "2026-08-14T00:00:02Z", data: { delta: "hel" } },
    ],
    // After Last-Event-ID=3 reconnect the server re-plays 3 (last seen)
    // and then continues with fresh frames; the hook must dedupe 3.
    replay: [
      { id: 3, kind: "token" as const, run_id: "run-1", at: "2026-08-14T00:00:02Z", data: { delta: "hel" } },
      { id: 4, kind: "token" as const, run_id: "run-1", at: "2026-08-14T00:00:03Z", data: { delta: "lo" } },
      { id: 5, kind: "done" as const, run_id: "run-1", at: "2026-08-14T00:00:04Z", data: { status: "done" as const } },
    ],
  };
}

/** Build a session-event frame carrying the current schema version. */
export function draftAssistSessionFrame(
  sessionId: string,
  frameId = 1,
  snapshot: DraftAssistSnapshot = {},
) {
  return {
    id: frameId,
    kind: "session" as const,
    at: "2026-08-14T00:00:00Z",
    data: {
      session_id: sessionId,
      snapshot,
      schema_version: DRAFT_ASSIST_SCHEMA_VERSION,
    },
  };
}

/** DELETE /draft-assist/sessions/{id} — 204. */
export function draftAssistDeleteSession(sessionId: string) {
  return http.delete(`${draftAssistBase}/sessions/${sessionId}`, () =>
    new HttpResponse(null, { status: 204 }),
  );
}

/** DELETE /draft-assist/sessions/{id} — 404 (server already GC-ed). */
export function draftAssistDeleteSessionNotFound(sessionId: string) {
  return http.delete(`${draftAssistBase}/sessions/${sessionId}`, () =>
    HttpResponse.json({ error: "not found" }, { status: 404 }),
  );
}

/** DELETE /draft-assist/sessions/{id} — capture calls for lifecycle tests. */
export function draftAssistDeleteSessionCapture(
  sessionId: string,
  onDelete: () => void,
) {
  return http.delete(`${draftAssistBase}/sessions/${sessionId}`, () => {
    onDelete();
    return new HttpResponse(null, { status: 204 });
  });
}

/** Convenience: all handlers a full happy-path composer test needs. */
export function draftAssistHappyPathHandlers(
  fixture: DraftAssistSessionFixture = makeDraftAssistSessionFixture(),
  runId = "run-1",
) {
  return [
    draftAssistReadyOk("fake"),
    draftAssistCreateSession(fixture),
    draftAssistGetSession(fixture),
    draftAssistUpdateSnapshot(fixture.id, fixture.snapshot),
    draftAssistStartRunOk(fixture.id, runId),
    draftAssistCancelRunOk(fixture.id, runId),
    draftAssistDeleteSession(fixture.id),
  ];
}
