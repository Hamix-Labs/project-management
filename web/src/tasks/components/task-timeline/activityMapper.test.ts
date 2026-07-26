import { describe, expect, it } from "vitest";
import {
  mapActivityEventToTimeline,
  mapActivityEventsToTimeline,
} from "./activityMapper";
import type { TaskActivityEvent } from "@/types";

const BASE: Omit<TaskActivityEvent, "type" | "data"> = {
  task_id: "f0000131-0000-4000-8000-000000000131",
  seq: 7,
  at: "2026-07-25T10:00:00.000Z",
  by: "user",
  // Timeline never renders `task_title` — kept on the envelope to
  // pin the "highlight comes from context, not title" contract.
  task_title: "My test task",
  task_number: 42,
  task_priority: "high",
  task_project_id: "proj-1",
  task_tags: ["api"],
};

describe("mapActivityEventToTimeline — status_changed", () => {
  it("shows #N once as taskRef and leaves highlight empty", () => {
    const ev: TaskActivityEvent = {
      ...BASE,
      type: "status_changed",
      data: { from: "ready", to: "running" },
    };
    const result = mapActivityEventToTimeline(ev);
    expect(result).not.toBeNull();
    expect(result!.kind).toBe("status-changed");
    expect(result!.category).toBe("tasks");
    expect(result!.title).toBe("Status changed");
    expect(result!.highlight).toBe("");
    expect(result!.taskRef).toBe("#42");
    expect(result!.detail).toBe("Ready → Running");
    expect(result!.meta).toBeUndefined();
    expect(result!.seq).toBe(7);
    expect(result!.taskId).toBe("f0000131-0000-4000-8000-000000000131");
    expect(result!.taskPriority).toBe("high");
    expect(result!.taskProjectId).toBe("proj-1");
    expect(result!.taskTags).toEqual(["api"]);
  });

  it("falls back to the shortened UUID when task_number is missing", () => {
    const ev: TaskActivityEvent = {
      ...BASE,
      task_number: undefined,
      type: "status_changed",
      data: { from: "running", to: "review" },
    };
    const result = mapActivityEventToTimeline(ev);
    expect(result!.highlight).toBe("");
    expect(result!.taskRef).toBe("f0000131");
  });

  it("never surfaces task_title on the timeline card", () => {
    const ev: TaskActivityEvent = {
      ...BASE,
      task_number: undefined,
      type: "status_changed",
      data: { from: "ready", to: "running" },
    };
    const result = mapActivityEventToTimeline(ev);
    expect(result!.highlight).not.toContain("My test task");
    expect(result!.detail).not.toContain("My test task");
    expect(result!.taskTitle).toBe("My test task");
  });

  it("handles missing from/to gracefully", () => {
    const ev: TaskActivityEvent = {
      ...BASE,
      type: "status_changed",
      data: {},
    };
    const result = mapActivityEventToTimeline(ev);
    expect(result).not.toBeNull();
    expect(result!.kind).toBe("status-changed");
    expect(result!.detail).toBe("Status updated.");
    expect(result!.meta).toBeUndefined();
  });
});

describe("mapActivityEventToTimeline — phase_failed", () => {
  it("maps execute phase failure with failure_summary and #N ref once", () => {
    const ev: TaskActivityEvent = {
      ...BASE,
      type: "phase_failed",
      data: { phase: "execute", failure_summary: "Runner exited with code 1" },
    };
    const result = mapActivityEventToTimeline(ev);
    expect(result).not.toBeNull();
    expect(result!.kind).toBe("verification-failed");
    expect(result!.category).toBe("verification");
    expect(result!.title).toBe("Execute phase failed");
    expect(result!.detail).toBe("Runner exited with code 1");
    expect(result!.highlight).toBe("");
    expect(result!.taskRef).toBe("#42");
    expect(result!.meta).toBeUndefined();
    expect(result!.seq).toBe(7);
  });

  it("prefers details.standardized_message over summary", () => {
    const ev: TaskActivityEvent = {
      ...BASE,
      type: "phase_failed",
      data: {
        phase: "execute",
        summary: "Short title",
        details: {
          standardized_message: "Cursor account usage limit reached.",
          failure_kind: "cursor_usage_limit",
        },
      },
    };
    const result = mapActivityEventToTimeline(ev);
    expect(result!.detail).toBe("Cursor account usage limit reached.");
  });

  it("maps verify phase failure with verification snapshot", () => {
    const ev: TaskActivityEvent = {
      ...BASE,
      type: "phase_failed",
      data: {
        phase: "verify",
        details: {
          verification: {
            attempt_seq: 2,
            passed_count: 1,
            failed_count: 2,
            criteria: [
              { criterion_id: "c1", verified: true },
              { criterion_id: "c2", verified: false },
              { criterion_id: "c3", verified: false },
            ],
          },
        },
      },
    };
    const result = mapActivityEventToTimeline(ev);
    expect(result).not.toBeNull();
    expect(result!.title).toBe("Verification failed");
    expect(result!.highlight).toBe("c2");
    expect(result!.detail).toBe("2 criteria failed");
    expect(result!.meta).toContain("1 passed");
    expect(result!.meta).toContain("2 failed");
  });
});

describe("mapActivityEventToTimeline — approval_granted", () => {
  it("maps approval with empty highlight and taskRef", () => {
    const ev: TaskActivityEvent = {
      ...BASE,
      type: "approval_granted",
      data: {},
    };
    const result = mapActivityEventToTimeline(ev);
    expect(result!.kind).toBe("review-approved");
    expect(result!.highlight).toBe("");
    expect(result!.taskRef).toBe("#42");
  });
});

describe("mapActivityEventsToTimeline", () => {
  it("drops unknown types", () => {
    const events: TaskActivityEvent[] = [
      {
        ...BASE,
        type: "status_changed",
        data: { from: "ready", to: "running" },
      },
    ];
    expect(mapActivityEventsToTimeline(events)).toHaveLength(1);
  });
});
