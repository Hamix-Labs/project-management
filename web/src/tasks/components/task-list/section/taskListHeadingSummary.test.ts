import { describe, expect, it } from "vitest";
import type { TaskStatsResponse } from "@/types";
import { formatTaskListHeadingSummary } from "./taskListHeadingSummary";

function makeStats(
  overrides: Partial<TaskStatsResponse> = {},
): TaskStatsResponse {
  return {
    total: 0,
    ready: 0,
    critical: 0,
    scheduled: 0,
    by_status: {},
    by_priority: {},
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
    ...overrides,
  };
}

describe("formatTaskListHeadingSummary", () => {
  it("returns undefined when nothing to show", () => {
    expect(formatTaskListHeadingSummary(0, null)).toBeUndefined();
    expect(formatTaskListHeadingSummary(0, makeStats())).toBeUndefined();
  });

  it("formats shown count with stats segments", () => {
    expect(
      formatTaskListHeadingSummary(
        15,
        makeStats({
          total: 15,
          ready: 7,
          by_status: { review: 2, blocked: 2 },
        }),
      ),
    ).toBe("15 shown · 7 ready · 2 in review · 2 blocked");
  });

  it("omits zero stat segments", () => {
    expect(
      formatTaskListHeadingSummary(
        3,
        makeStats({
          total: 3,
          ready: 3,
        }),
      ),
    ).toBe("3 shown · 3 ready");
  });

  it("shows zero shown when filters hide all rows but stats exist", () => {
    expect(
      formatTaskListHeadingSummary(
        0,
        makeStats({
          total: 5,
          ready: 5,
        }),
      ),
    ).toBe("0 shown · 5 ready");
  });
});
