import { describe, expect, it } from "vitest";
import { DRAFT_ASSIST_SCHEMA_VERSION } from "@/types/draftAssist";
import {
  parseDraftAssistCancelRunResult,
  parseDraftAssistEvent,
  parseDraftAssistReady,
  parseDraftAssistSession,
  parseDraftAssistSessionEventData,
  parseDraftAssistSnapshot,
  parseDraftAssistSnapshotUpdate,
  parseDraftAssistStartRunResult,
} from "./parseTaskApiDraftAssist";

const sessionWire = {
  id: "sess-1",
  nonce: "nonce-1",
  worktree_id: "wt-1",
  snapshot: {
    title: "T",
    prompt: "P",
    criteria: ["c1"],
    tags: ["tag"],
    cursor_model: "auto",
    updated_at: "2026-08-14T00:00:00Z",
  },
  created_at: "2026-08-14T00:00:00Z",
  updated_at: "2026-08-14T00:00:01Z",
};

describe("draft-assist REST parsers", () => {
  it("round-trips a session", () => {
    const s = parseDraftAssistSession(sessionWire);
    expect(s.id).toBe("sess-1");
    expect(s.nonce).toBe("nonce-1");
    expect(s.worktree_id).toBe("wt-1");
    expect(s.snapshot.title).toBe("T");
    expect(s.snapshot.criteria).toEqual(["c1"]);
    expect(s.updated_at).toBe("2026-08-14T00:00:01Z");
  });

  it("accepts an empty snapshot", () => {
    expect(parseDraftAssistSnapshot({})).toEqual({});
    expect(parseDraftAssistSnapshot(undefined)).toEqual({});
  });

  it("rejects a snapshot with non-string criteria", () => {
    expect(() => parseDraftAssistSnapshot({ criteria: [1] })).toThrow(
      /snapshot.criteria\[0\] must be a string/,
    );
  });

  it("throws when the session lacks a nonce", () => {
    expect(() =>
      parseDraftAssistSession({ ...sessionWire, nonce: "" }),
    ).toThrow(/nonce/);
  });

  it("parses a snapshot update", () => {
    expect(
      parseDraftAssistSnapshotUpdate({ id: "sess-1", snapshot: { prompt: "x" } }),
    ).toEqual({ id: "sess-1", snapshot: { prompt: "x" } });
  });

  it("parses a start-run result", () => {
    expect(parseDraftAssistStartRunResult({ run_id: "r1" })).toEqual({
      run_id: "r1",
    });
  });

  it("parses a cancel-run result and rejects wrong status", () => {
    expect(
      parseDraftAssistCancelRunResult({ run_id: "r1", status: "cancelling" }),
    ).toEqual({ run_id: "r1", status: "cancelling" });
    expect(() =>
      parseDraftAssistCancelRunResult({ run_id: "r1", status: "done" }),
    ).toThrow(/cancel status/);
  });

  it("parses a ready probe with a reason", () => {
    expect(
      parseDraftAssistReady({
        ready: false,
        runner: "missing",
        reason: "no_runner",
      }),
    ).toEqual({ ready: false, runner: "missing", reason: "no_runner" });
    expect(parseDraftAssistReady({ ready: true, runner: "fake" })).toEqual({
      ready: true,
      runner: "fake",
    });
  });

  it("rejects unknown ready reasons", () => {
    expect(() =>
      parseDraftAssistReady({ ready: false, runner: "sdk", reason: "meh" }),
    ).toThrow(/reason must be one of/);
  });
});

describe("draft-assist SSE event parser", () => {
  it("parses a session envelope and enforces schema_version", () => {
    const ev = parseDraftAssistEvent({
      id: 1,
      kind: "session",
      at: "2026-08-14T00:00:00Z",
      data: {
        session_id: "sess-1",
        snapshot: { prompt: "hi" },
        schema_version: DRAFT_ASSIST_SCHEMA_VERSION,
      },
    });
    expect(ev.kind).toBe("session");
    if (ev.kind === "session") {
      expect(ev.data.session_id).toBe("sess-1");
      expect(ev.data.snapshot.prompt).toBe("hi");
    }
  });

  it("hard-throws on schema_version mismatch", () => {
    expect(() =>
      parseDraftAssistSessionEventData({
        session_id: "sess-1",
        snapshot: {},
        schema_version: DRAFT_ASSIST_SCHEMA_VERSION + 1,
      }),
    ).toThrow(/schema mismatch/);
  });

  it("parses status/token/tool/patch/error/done frames", () => {
    const at = "2026-08-14T00:00:01Z";
    expect(
      parseDraftAssistEvent({
        id: 2,
        kind: "status",
        run_id: "r1",
        at,
        data: { status: "thinking" },
      }).kind,
    ).toBe("status");
    expect(
      parseDraftAssistEvent({
        id: 3,
        kind: "token",
        run_id: "r1",
        at,
        data: { delta: "hi" },
      }).kind,
    ).toBe("token");
    expect(
      parseDraftAssistEvent({
        id: 4,
        kind: "tool",
        run_id: "r1",
        at,
        data: { name: "draft_set_prompt", phase: "start" },
      }).kind,
    ).toBe("tool");
    expect(
      parseDraftAssistEvent({
        id: 5,
        kind: "patch",
        run_id: "r1",
        at,
        data: { op: "set", value: "x" },
      }).kind,
    ).toBe("patch");
    expect(
      parseDraftAssistEvent({
        id: 6,
        kind: "error",
        run_id: "r1",
        at,
        data: { code: "runner_error", message: "boom" },
      }).kind,
    ).toBe("error");
    expect(
      parseDraftAssistEvent({
        id: 7,
        kind: "done",
        run_id: "r1",
        at,
        data: { status: "done" },
      }).kind,
    ).toBe("done");
  });

  it("rejects unknown event kinds", () => {
    expect(() =>
      parseDraftAssistEvent({
        id: 8,
        kind: "unknown",
        at: "2026-08-14T00:00:01Z",
        data: {},
      }),
    ).toThrow(/kind must be one of/);
  });

  it("rejects a bad tool phase", () => {
    expect(() =>
      parseDraftAssistEvent({
        id: 9,
        kind: "tool",
        at: "2026-08-14T00:00:01Z",
        data: { name: "x", phase: "middle" },
      }),
    ).toThrow(/tool.phase/);
  });
});
