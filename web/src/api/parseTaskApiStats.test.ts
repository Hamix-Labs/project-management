import { describe, expect, it } from "vitest";
import {
  parseCycleFailuresListResponse,
  parseTaskStatsResponse,
} from "./parseTaskApiStats";

describe("parseTaskStatsResponse", () => {
  // emptyExtras covers the cycle/phase/recent_failures blocks the
  // server always sends — most assertions in this suite focus on the
  // task-counter half of the envelope and reuse this stub.
  const emptyExtras = {
    cycles: { by_status: {}, by_triggered_by: {} },
    phases: {
      by_phase_status: {
        execute: {},
        verify: {},
      },
    },
    runner: {
      by_runner: {},
      by_model: {},
      by_runner_model: {},
      by_runner_model_resolved: {},
    },
    recent_failures: [],
  };

  it("parses task stats envelope", () => {
    expect(
      parseTaskStatsResponse({
        total: 22,
        ready: 7,
        critical: 2,
        scheduled: 3,
        by_status: { ready: 7, running: 5 },
        by_priority: { critical: 2, high: 4 },
        ...emptyExtras,
      }),
    ).toEqual({
      total: 22,
      ready: 7,
      critical: 2,
      scheduled: 3,
      by_status: { ready: 7, running: 5 },
      by_priority: { critical: 2, high: 4 },
      ...emptyExtras,
    });
  });

  it("defaults scheduled to 0 when omitted (back-compat with pre-Stage-6 backends)", () => {
    const got = parseTaskStatsResponse({
      total: 1,
      ready: 1,
      critical: 0,
      // scheduled key intentionally absent — older backend
      by_status: { ready: 1 },
      by_priority: {},
      ...emptyExtras,
    });
    expect(got.scheduled).toBe(0);
  });

  it("rejects scheduled when present-but-non-numeric", () => {
    expect(() =>
      parseTaskStatsResponse({
        total: 1,
        ready: 1,
        critical: 0,
        scheduled: "3",
        by_status: { ready: 1 },
        by_priority: {},
        ...emptyExtras,
      }),
    ).toThrow(/scheduled/);
  });

  it("rejects invalid stats payload", () => {
    expect(() =>
      parseTaskStatsResponse({
        total: "22",
        ready: 7,
        critical: 2,
        by_status: {},
        by_priority: {},
        ...emptyExtras,
      }),
    ).toThrow(/total/);
  });

  it("rejects unknown status/priority keys in breakdowns", () => {
    expect(() =>
      parseTaskStatsResponse({
        total: 22,
        ready: 7,
        critical: 2,
        by_status: { nope: 1 },
        by_priority: {},
        ...emptyExtras,
      }),
    ).toThrow(/known status/);

    expect(() =>
      parseTaskStatsResponse({
        total: 22,
        ready: 7,
        critical: 2,
        by_status: {},
        by_priority: { urgent: 1 },
        ...emptyExtras,
      }),
    ).toThrow(/known priority/);
  });

  it("parses cycles aggregates and rejects unknown enums", () => {
    const got = parseTaskStatsResponse({
      total: 0,
      ready: 0,
      critical: 0,
      by_status: {},
      by_priority: {},
      ...emptyExtras,
      cycles: {
        by_status: { running: 1, succeeded: 4, failed: 2, aborted: 1 },
        by_triggered_by: { user: 3, agent: 5 },
      },
    });
    expect(got.cycles.by_status).toEqual({
      running: 1,
      succeeded: 4,
      failed: 2,
      aborted: 1,
    });
    expect(got.cycles.by_triggered_by).toEqual({ user: 3, agent: 5 });

    expect(() =>
      parseTaskStatsResponse({
        total: 0,
        ready: 0,
        critical: 0,
        by_status: {},
        by_priority: {},
        ...emptyExtras,
        cycles: {
          by_status: { weird: 1 },
          by_triggered_by: {},
        },
      }),
    ).toThrow(/cycles\.by_status/);
  });

  // Legacy phase buckets (diagnose / persist) can still appear on the
  // stats endpoint when historical task_cycle_phases rows exist. The
  // parser must drop them silently rather than throwing — otherwise the
  // Observability page breaks the moment a deprecated phase value is
  // returned by an older deployment.
  it("ignores legacy diagnose / persist buckets in stats response", () => {
    const got = parseTaskStatsResponse({
      total: 0,
      ready: 0,
      critical: 0,
      by_status: {},
      by_priority: {},
      ...emptyExtras,
      phases: {
        by_phase_status: {
          execute: { succeeded: 3 },
          verify: {},
          diagnose: { skipped: 9 },
          persist: { succeeded: 1 },
        },
      },
    });
    expect(Object.keys(got.phases.by_phase_status).sort()).toEqual([
      "execute",
      "verify",
    ]);
    expect(got.phases.by_phase_status.execute).toEqual({ succeeded: 3 });
  });

  it("parses phases heatmap with all writable phases always present", () => {
    const got = parseTaskStatsResponse({
      total: 0,
      ready: 0,
      critical: 0,
      by_status: {},
      by_priority: {},
      ...emptyExtras,
      phases: {
        by_phase_status: {
          // Server omits phases with no data; parser must still seed
          // every Phase enum value with `{}` so the heatmap renders.
          execute: { failed: 2, succeeded: 1 },
          verify: {},
        },
      },
    });
    expect(Object.keys(got.phases.by_phase_status).sort()).toEqual([
      "execute",
      "verify",
    ]);
    expect(got.phases.by_phase_status.execute).toEqual({
      failed: 2,
      succeeded: 1,
    });
  });

  it("parses recent_failures and rejects bad rows", () => {
    const got = parseTaskStatsResponse({
      total: 0,
      ready: 0,
      critical: 0,
      by_status: {},
      by_priority: {},
      ...emptyExtras,
      recent_failures: [
        {
          task_id: "t-1",
          event_seq: 7,
          at: "2026-04-19T12:00:00Z",
          cycle_id: "c-1",
          attempt_seq: 2,
          status: "failed",
          reason: "execute blew up",
        },
      ],
    });
    expect(got.recent_failures).toHaveLength(1);
    expect(got.recent_failures[0].cycle_id).toBe("c-1");

    expect(() =>
      parseTaskStatsResponse({
        total: 0,
        ready: 0,
        critical: 0,
        by_status: {},
        by_priority: {},
        ...emptyExtras,
        recent_failures: [
          {
            task_id: "t-1",
            event_seq: 7,
            at: "2026-04-19T12:00:00Z",
            cycle_id: "c-1",
            attempt_seq: 2,
            status: "succeeded",
            reason: "",
          },
        ],
      }),
    ).toThrow(/recent_failures\[0\]\.status/);
  });

  it("parses cycle failures list responses", () => {
    const got = parseCycleFailuresListResponse({
      total: 1,
      limit: 50,
      offset: 0,
      sort: "at_desc",
      reason_sort_truncated: false,
      failures: [
        {
          task_id: "t-1",
          event_seq: 7,
          at: "2026-04-19T12:00:00Z",
          cycle_id: "c-1",
          attempt_seq: 2,
          status: "failed",
          reason: "execute blew up",
        },
      ],
    });
    expect(got.total).toBe(1);
    expect(got.sort).toBe("at_desc");
    expect(got.reason_sort_truncated).toBe(false);
    expect(got.failures).toHaveLength(1);
    expect(got.failures[0].task_id).toBe("t-1");
  });

  it("parses runner breakdown across by_runner, by_model, and by_runner_model", () => {
    const got = parseTaskStatsResponse({
      total: 0,
      ready: 0,
      critical: 0,
      by_status: {},
      by_priority: {},
      ...emptyExtras,
      runner: {
        by_runner: {
          "cursor-cli": {
            by_status: { succeeded: 2, failed: 1 },
            succeeded: 2,
            duration_p50_succeeded_seconds: 1.5,
            duration_p95_succeeded_seconds: 4.2,
          },
        },
        by_model: {
          "sonnet-4.5": {
            by_status: { succeeded: 1 },
            succeeded: 1,
            duration_p50_succeeded_seconds: 1,
            duration_p95_succeeded_seconds: 1,
          },
          "": {
            by_status: { failed: 1 },
            succeeded: 0,
            duration_p50_succeeded_seconds: 0,
            duration_p95_succeeded_seconds: 0,
          },
        },
        by_runner_model: {
          "cursor-cli|sonnet-4.5": {
            by_status: { succeeded: 1 },
            succeeded: 1,
            duration_p50_succeeded_seconds: 1,
            duration_p95_succeeded_seconds: 1,
          },
        },
      },
    });
    expect(got.runner.by_runner["cursor-cli"].succeeded).toBe(2);
    expect(got.runner.by_model[""].by_status.failed).toBe(1);
    expect(got.runner.by_runner_model["cursor-cli|sonnet-4.5"].duration_p95_succeeded_seconds).toBe(1);
    // Older backends predate by_runner_model_resolved on the wire.
    // Parser must tolerate the missing key by defaulting to `{}` so
    // the SPA doesn't hard-error during a rollout.
    expect(got.runner.by_runner_model_resolved).toEqual({});
  });

  it("parses by_runner_model_resolved when the server emits it", () => {
    const got = parseTaskStatsResponse({
      total: 0,
      ready: 0,
      critical: 0,
      by_status: {},
      by_priority: {},
      ...emptyExtras,
      runner: {
        by_runner: {},
        by_model: {},
        by_runner_model: {},
        by_runner_model_resolved: {
          "cursor-cli|auto|claude-4-sonnet": {
            by_status: { succeeded: 3 },
            succeeded: 3,
            duration_p50_succeeded_seconds: 2,
            duration_p95_succeeded_seconds: 7,
          },
        },
      },
    });
    expect(
      got.runner.by_runner_model_resolved["cursor-cli|auto|claude-4-sonnet"]
        .succeeded,
    ).toBe(3);
  });

  it("rejects unknown cycle status keys inside a runner bucket", () => {
    expect(() =>
      parseTaskStatsResponse({
        total: 0,
        ready: 0,
        critical: 0,
        by_status: {},
        by_priority: {},
        ...emptyExtras,
        runner: {
          by_runner: {
            "cursor-cli": {
              by_status: { weird: 1 },
              succeeded: 0,
              duration_p50_succeeded_seconds: 0,
              duration_p95_succeeded_seconds: 0,
            },
          },
          by_model: {},
          by_runner_model: {},
        },
      }),
    ).toThrow(/by_status\.weird/);
  });

  it("rejects missing runner block entirely", () => {
    const { runner: _omit, ...withoutRunner } = emptyExtras;
    expect(() =>
      parseTaskStatsResponse({
        total: 0,
        ready: 0,
        critical: 0,
        by_status: {},
        by_priority: {},
        ...withoutRunner,
      }),
    ).toThrow(/runner must be an object/);
  });
});
