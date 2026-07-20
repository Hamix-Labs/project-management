import { describe, expect, it } from "vitest";
import {
  parseCycleLifecycleEventData,
  parsePhaseLifecycleEventData,
  parseTaskEventData,
  parseTransitionEventData,
} from "./parseTaskApiEventData";
import { parseTaskEventDetail, parseTaskEventsResponse } from "./parseTaskApiEvents";

describe("parseTaskEventData families", () => {
  it("parses phase lifecycle fields with typed numbers/strings", () => {
    expect(
      parsePhaseLifecycleEventData({
        phase: "execute",
        status: "succeeded",
        phase_seq: 2,
        cycle_id: "c1",
        summary: "ok",
        details: { duration_ms: 10 },
      }),
    ).toEqual({
      phase: "execute",
      status: "succeeded",
      phase_seq: 2,
      cycle_id: "c1",
      summary: "ok",
      details: { duration_ms: 10 },
    });
  });

  it("throws when phase_seq is not a number", () => {
    expect(() =>
      parsePhaseLifecycleEventData({ phase: "execute", phase_seq: "2" }),
    ).toThrow(/phase_seq must be a number/);
  });

  it("throws when details is not an object", () => {
    expect(() =>
      parsePhaseLifecycleEventData({ details: "nope" }),
    ).toThrow(/details must be an object/);
  });

  it("parses cycle terminal failure fields", () => {
    expect(
      parseCycleLifecycleEventData({
        cycle_id: "c1",
        attempt_seq: 1,
        status: "failed",
        reason: "x",
        failure_summary: "boom",
      }),
    ).toEqual({
      cycle_id: "c1",
      attempt_seq: 1,
      status: "failed",
      reason: "x",
      failure_summary: "boom",
    });
  });

  it("parses transition from/to", () => {
    expect(parseTransitionEventData({ from: "ready", to: "running" })).toEqual({
      from: "ready",
      to: "running",
    });
  });

  it("routes by event type", () => {
    expect(parseTaskEventData("phase_failed", { phase: "verify" })).toEqual({
      phase: "verify",
    });
    expect(parseTaskEventData("task_created", { title: "T" })).toEqual({
      title: "T",
    });
  });

  it("throws when data is an array", () => {
    expect(() => parseTaskEventData("sync_ping", [])).toThrow(
      /event data must be an object/,
    );
  });
});

describe("parseTaskEventsResponse with typed data", () => {
  it("parses a phase_completed list row", () => {
    const parsed = parseTaskEventsResponse({
      task_id: "t1",
      events: [
        {
          seq: 1,
          at: "2026-01-01T00:00:00Z",
          type: "phase_completed",
          by: "agent",
          data: { phase: "execute", status: "succeeded", phase_seq: 1 },
        },
      ],
      approval_pending: false,
    });
    expect(parsed.events[0]?.type).toBe("phase_completed");
    expect(parsed.events[0]?.data).toEqual({
      phase: "execute",
      status: "succeeded",
      phase_seq: 1,
    });
  });

  it("rejects malformed phase_seq in detail", () => {
    expect(() =>
      parseTaskEventDetail({
        task_id: "t1",
        seq: 1,
        at: "2026-01-01T00:00:00Z",
        type: "phase_failed",
        by: "agent",
        data: { phase: "verify", phase_seq: true },
      }),
    ).toThrow(/phase_seq/);
  });
});
