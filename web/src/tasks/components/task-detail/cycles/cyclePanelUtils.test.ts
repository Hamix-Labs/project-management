import { describe, expect, it } from "vitest";
import type { TaskCycle, TaskCyclePhase, TaskCyclesListResponse } from "@/types/cycle";
import {
  formatAttemptTiming,
  indexCyclesById,
  pickLatestPhase,
  pickRunningPhase,
  splitRunningAndHistory,
} from "./cyclePanelUtils";

const emptyCycleMeta = {
  runner: "",
  runner_version: "",
  cursor_model: "",
  cursor_model_effective: "",
  prompt_hash: "",
};

function cycle(id: string, status: TaskCycle["status"]): TaskCycle {
  return {
    id,
    task_id: "task-1",
    status,
    attempt_seq: 1,
    triggered_by: "user",
    started_at: "2026-01-01T12:00:00.000Z",
    meta: {},
    cycle_meta: emptyCycleMeta,
  };
}

function phase(
  id: string,
  phaseName: TaskCyclePhase["phase"],
  status: TaskCyclePhase["status"],
  phaseSeq: number,
): TaskCyclePhase {
  return {
    id,
    cycle_id: "cyc-1",
    phase: phaseName,
    phase_seq: phaseSeq,
    status,
    started_at: "2026-01-01T12:00:00.000Z",
    details: {},
  };
}

describe("splitRunningAndHistory", () => {
  it("returns empty when envelope is undefined", () => {
    expect(splitRunningAndHistory(undefined)).toEqual({
      runningCycle: null,
      historyCycles: [],
    });
  });

  it("surfaces the running cycle separately while history includes all rows", () => {
    const running = cycle("c-run", "running");
    const done = cycle("c-done", "succeeded");
    const envelope: TaskCyclesListResponse = {
      task_id: "task-1",
      limit: 20,
      cycles: [running, done],
      has_more: false,
    };
    expect(splitRunningAndHistory(envelope)).toEqual({
      runningCycle: running,
      historyCycles: [running, done],
    });
  });
});

describe("indexCyclesById", () => {
  it("maps cycle ids to rows", () => {
    const a = cycle("a", "succeeded");
    const b = cycle("b", "failed");
    const map = indexCyclesById([a, b]);
    expect(map.get("a")).toBe(a);
    expect(map.get("b")).toBe(b);
  });
});

describe("pickRunningPhase", () => {
  it("returns the running phase when present", () => {
    const phases = [
      phase("p1", "execute", "succeeded", 1),
      phase("p2", "verify", "running", 2),
    ];
    expect(pickRunningPhase(phases)?.id).toBe("p2");
  });
});

describe("pickLatestPhase", () => {
  it("returns the highest phase_seq", () => {
    const phases = [
      phase("p1", "execute", "succeeded", 1),
      phase("p2", "verify", "succeeded", 2),
    ];
    expect(pickLatestPhase(phases)?.id).toBe("p2");
  });
});

describe("formatAttemptTiming", () => {
  it("labels in-progress attempts without a duration", () => {
    const timing = formatAttemptTiming(cycle("c-run", "running"), "UTC");
    expect(timing.label).toMatch(/^Picked up · .+ · Completed · in progress$/);
    expect(timing.label).not.toMatch(/min/);
  });

  it("labels terminal attempts with completed time and duration", () => {
    const timing = formatAttemptTiming(
      {
        ...cycle("c-done", "succeeded"),
        started_at: "2026-04-22T12:48:00Z",
        ended_at: "2026-04-22T13:00:00Z",
      },
      "UTC",
    );
    expect(timing.label).toMatch(/^Picked up · .+ · Completed · .+ · 12 min$/);
    expect(timing.ariaLabel).toBe(timing.label);
  });
});
