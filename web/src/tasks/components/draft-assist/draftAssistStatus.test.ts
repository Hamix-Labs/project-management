import { describe, expect, it } from "vitest";
import type { DraftAssistEvent } from "@/types/draftAssist";
import {
  draftAssistStatusCopy,
  draftAssistStatusReducer,
  INITIAL_DRAFT_ASSIST_STATUS,
  isDraftAssistRunActive,
  type DraftAssistStatusState,
} from "./draftAssistStatus";

const at = "2026-08-14T00:00:00Z";

function event(kind: DraftAssistEvent["kind"], data: unknown, id = 1): DraftAssistEvent {
  return {
    id,
    kind,
    at,
    run_id: "run-1",
    data,
  } as DraftAssistEvent;
}

function reduceAll(actions: Parameters<typeof draftAssistStatusReducer>[1][]): DraftAssistStatusState {
  return actions.reduce(draftAssistStatusReducer, INITIAL_DRAFT_ASSIST_STATUS);
}

describe("draftAssistStatusReducer", () => {
  it("starts idle with empty copy", () => {
    expect(INITIAL_DRAFT_ASSIST_STATUS.status).toBe("idle");
    expect(draftAssistStatusCopy(INITIAL_DRAFT_ASSIST_STATUS)).toBe("");
    expect(isDraftAssistRunActive(INITIAL_DRAFT_ASSIST_STATUS)).toBe(false);
  });

  it("run_requested → starting; copy matches design", () => {
    const s = draftAssistStatusReducer(INITIAL_DRAFT_ASSIST_STATUS, {
      type: "run_requested",
    });
    expect(s.status).toBe("starting");
    expect(draftAssistStatusCopy(s)).toBe("Starting assistant…");
    expect(isDraftAssistRunActive(s)).toBe(true);
  });

  it("status frames advance through thinking → streaming → applying", () => {
    const s = reduceAll([
      { type: "run_requested" },
      { type: "event", event: event("status", { status: "thinking" }) },
    ]);
    expect(s.status).toBe("thinking");
    expect(draftAssistStatusCopy(s)).toBe("Thinking…");

    const s2 = draftAssistStatusReducer(s, {
      type: "event",
      event: event("token", { delta: "hi" }, 2),
    });
    expect(s2.status).toBe("streaming");

    const s3 = draftAssistStatusReducer(s2, {
      type: "event",
      event: event("patch", { op: "set", value: "<p>x</p>", summary: "tightened" }, 3),
    });
    expect(s3.status).toBe("applying");
    expect(draftAssistStatusCopy(s3)).toBe("Updating prompt — tightened");
  });

  it("tool start surfaces the name in the copy; tool end drops it", () => {
    const s = reduceAll([
      { type: "run_requested" },
      { type: "event", event: event("tool", { name: "hamix.draft_read_file", phase: "start" }) },
    ]);
    expect(s.status).toBe("tool");
    expect(s.toolName).toBe("hamix.draft_read_file");
    expect(draftAssistStatusCopy(s)).toBe("Reading hamix.draft_read_file…");

    const s2 = draftAssistStatusReducer(s, {
      type: "event",
      event: event("tool", { name: "hamix.draft_read_file", phase: "end", ok: true }, 2),
    });
    expect(s2.toolName).toBeNull();
    expect(s2.status).toBe("thinking");
  });

  it("cancel_requested → cancelling, done{cancelled} → terminal Assistant stopped", () => {
    const s = reduceAll([
      { type: "run_requested" },
      { type: "event", event: event("status", { status: "streaming" }) },
      { type: "cancel_requested" },
    ]);
    expect(s.status).toBe("cancelling");
    expect(draftAssistStatusCopy(s)).toBe("Stopping…");

    const s2 = draftAssistStatusReducer(s, {
      type: "event",
      event: event("done", { status: "cancelled" }, 2),
    });
    expect(s2.status).toBe("idle");
    expect(s2.terminal).toBe("assistant_stopped");
    expect(draftAssistStatusCopy(s2)).toBe("Assistant stopped");
  });

  it("done after patch → Prompt updated terminal note", () => {
    const s = reduceAll([
      { type: "run_requested" },
      { type: "event", event: event("patch", { op: "set", value: "<p>x</p>", summary: "s" }, 1) },
      { type: "event", event: event("done", { status: "done" }, 2) },
    ]);
    expect(s.status).toBe("idle");
    expect(s.terminal).toBe("prompt_updated");
    expect(draftAssistStatusCopy(s)).toBe("Prompt updated");
  });

  it("connection loss during run → disconnected; recovery goes back to thinking", () => {
    const s = reduceAll([
      { type: "run_requested" },
      { type: "event", event: event("status", { status: "streaming" }) },
      { type: "connection", connected: false },
    ]);
    expect(s.status).toBe("disconnected");
    expect(draftAssistStatusCopy(s)).toBe("Reconnecting…");

    const s2 = draftAssistStatusReducer(s, { type: "connection", connected: true });
    expect(s2.status).toBe("thinking");
  });

  it("connection loss while idle stays idle", () => {
    const s = draftAssistStatusReducer(INITIAL_DRAFT_ASSIST_STATUS, {
      type: "connection",
      connected: false,
    });
    expect(s.status).toBe("idle");
  });

  it("error event and done{failed} surface errorMessage", () => {
    const s = draftAssistStatusReducer(
      { ...INITIAL_DRAFT_ASSIST_STATUS, status: "streaming" },
      { type: "event", event: event("error", { code: "boom", message: "kaboom" }) },
    );
    expect(s.status).toBe("error");
    expect(s.errorMessage).toBe("kaboom");
    expect(draftAssistStatusCopy(s)).toContain("kaboom");

    const s2 = draftAssistStatusReducer(
      { ...INITIAL_DRAFT_ASSIST_STATUS, status: "streaming" },
      { type: "event", event: event("done", { status: "failed" }, 2) },
    );
    expect(s2.status).toBe("error");
    expect(s2.errorMessage).toBeTruthy();
  });

  it("transport_error surfaces message and enters error state", () => {
    const s = draftAssistStatusReducer(
      { ...INITIAL_DRAFT_ASSIST_STATUS, status: "starting" },
      { type: "transport_error", message: "timeout" },
    );
    expect(s.status).toBe("error");
    expect(s.errorMessage).toBe("timeout");
  });

  it("reset clears state back to idle", () => {
    const s = reduceAll([
      { type: "run_requested" },
      { type: "event", event: event("status", { status: "thinking" }) },
      { type: "reset" },
    ]);
    expect(s).toEqual(INITIAL_DRAFT_ASSIST_STATUS);
  });
});
