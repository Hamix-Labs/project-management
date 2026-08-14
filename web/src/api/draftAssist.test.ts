// @vitest-environment jsdom
import { describe, expect, it } from "vitest";
import {
  cancelRun,
  createSession,
  deleteSession,
  draftAssistEventsUrl,
  getSession,
  readyProbe,
  startRun,
  updateSnapshot,
} from "./draftAssist";
import { ApiError } from "./shared";
import { ensureMswListening } from "@/test/mswLifecycle";
import { server } from "@/test/server";
import {
  draftAssistCancelRunOk,
  draftAssistCreateSession,
  draftAssistCreateSessionCapture,
  draftAssistDeleteSession,
  draftAssistDeleteSessionNotFound,
  draftAssistGetSession,
  draftAssistReadyMissing,
  draftAssistReadyOk,
  draftAssistStartRunConflict,
  draftAssistStartRunOk,
  draftAssistUpdateSnapshot,
  makeDraftAssistSessionFixture,
} from "@/test/handlers/draftAssist";

ensureMswListening("error");

describe("draft-assist API client", () => {
  it("readyProbe returns runner + reason when not ready", async () => {
    server.use(draftAssistReadyMissing("no_runner"));
    await expect(readyProbe()).resolves.toEqual({
      ready: false,
      runner: "missing",
      reason: "no_runner",
    });
  });

  it("readyProbe returns bare ready when the runner is up", async () => {
    server.use(draftAssistReadyOk("fake"));
    await expect(readyProbe()).resolves.toEqual({ ready: true, runner: "fake" });
  });

  it("createSession POSTs snapshot + worktree_id and parses the created row", async () => {
    let body = "";
    const fixture = makeDraftAssistSessionFixture({ id: "sess-A" });
    server.use(
      draftAssistCreateSessionCapture((b) => {
        body = b;
      }, fixture),
    );
    const created = await createSession({
      worktree_id: "wt-9",
      snapshot: { title: "hi" },
    });
    expect(created.id).toBe("sess-A");
    expect(created.nonce).toBe(fixture.nonce);
    const parsed = JSON.parse(body) as {
      worktree_id: string;
      snapshot: { title?: string };
    };
    expect(parsed.worktree_id).toBe("wt-9");
    expect(parsed.snapshot.title).toBe("hi");
  });

  it("getSession fetches by id", async () => {
    const fixture = makeDraftAssistSessionFixture({ id: "sess-B" });
    server.use(draftAssistGetSession(fixture));
    const s = await getSession("sess-B");
    expect(s.id).toBe("sess-B");
  });

  it("updateSnapshot PUTs and parses the reply", async () => {
    const snap = { prompt: "new" };
    server.use(draftAssistUpdateSnapshot("sess-C", snap));
    await expect(updateSnapshot("sess-C", snap)).resolves.toEqual({
      id: "sess-C",
      snapshot: snap,
    });
  });

  it("startRun returns the run id (202)", async () => {
    server.use(draftAssistStartRunOk("sess-D", "run-42"));
    await expect(
      startRun("sess-D", { user_message: "hi" }),
    ).resolves.toEqual({ run_id: "run-42" });
  });

  it("startRun surfaces 409 as an ApiError so callers can branch on status", async () => {
    server.use(draftAssistStartRunConflict("sess-E"));
    await expect(
      startRun("sess-E", { user_message: "hi" }),
    ).rejects.toMatchObject({
      name: "ApiError",
      status: 409,
    });
    // Redundant but explicit: instanceof check for consumers.
    const err = await startRun("sess-E", { user_message: "hi" }).catch((e) => e);
    expect(err).toBeInstanceOf(ApiError);
  });

  it("cancelRun returns { run_id, status: 'cancelling' }", async () => {
    server.use(draftAssistCancelRunOk("sess-F", "run-1"));
    await expect(cancelRun("sess-F", "run-1")).resolves.toEqual({
      run_id: "run-1",
      status: "cancelling",
    });
  });

  it("deleteSession succeeds on 204", async () => {
    server.use(draftAssistDeleteSession("sess-G"));
    await expect(deleteSession("sess-G")).resolves.toBeUndefined();
  });

  it("deleteSession tolerates 404 (server GC-ed the session)", async () => {
    server.use(draftAssistDeleteSessionNotFound("sess-H"));
    await expect(deleteSession("sess-H")).resolves.toBeUndefined();
  });

  it("draftAssistEventsUrl escapes the session id", () => {
    expect(draftAssistEventsUrl("sess a/b")).toBe(
      "/draft-assist/sessions/sess%20a%2Fb/events",
    );
  });

  it("throws before fetching when session id is empty", async () => {
    await expect(getSession("  ")).rejects.toThrow(/session id is required/);
  });

  it("createSession happy path also works with the shared fixture handler", async () => {
    server.use(draftAssistCreateSession());
    const created = await createSession({ snapshot: {} });
    expect(created.id).toBe("da-sess-1");
  });
});
