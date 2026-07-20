import { describe, expect, it } from "vitest";
import { makeTask } from "@/test/taskDefaults";
import type { TaskListResponse } from "@/types";
import { mergePatchIntoTask, patchTaskInList } from "./optimisticTaskDetail";

const basePatch = {
  title: "Patched",
  initial_prompt: "<p>new</p>",
  status: "ready" as const,
  priority: "high" as const,
  cursor_model: "gpt",
};

describe("mergePatchIntoTask", () => {
  it("overwrites required patch fields", () => {
    const task = makeTask({
      id: "t1",
      title: "Old",
      initial_prompt: "old",
      status: "running",
      priority: "low",
      cursor_model: "",
    });
    expect(mergePatchIntoTask(task, basePatch)).toMatchObject({
      id: "t1",
      title: "Patched",
      initial_prompt: "<p>new</p>",
      status: "ready",
      priority: "high",
      cursor_model: "gpt",
    });
  });

  it("clears nullable fields when patch sends null", () => {
    const task = makeTask({
      id: "t1",
      project_id: "proj-1",
      milestone: "M1",
      pickup_not_before: "2026-07-01T00:00:00Z",
    });
    const merged = mergePatchIntoTask(task, {
      ...basePatch,
      project_id: null,
      milestone: null,
      pickup_not_before: null,
    });
    expect(merged.project_id).toBeUndefined();
    expect(merged.milestone).toBeUndefined();
    expect(merged.pickup_not_before).toBeUndefined();
  });

  it("preserves fields when patch omits them (undefined)", () => {
    const task = makeTask({
      id: "t1",
      project_id: "proj-1",
      tags: ["a"],
      milestone: "M1",
      project_context_item_ids: ["c1"],
      pickup_not_before: "2026-07-01T00:00:00Z",
    });
    const merged = mergePatchIntoTask(task, basePatch);
    expect(merged.project_id).toBe("proj-1");
    expect(merged.tags).toEqual(["a"]);
    expect(merged.milestone).toBe("M1");
    expect(merged.project_context_item_ids).toEqual(["c1"]);
    expect(merged.pickup_not_before).toBe("2026-07-01T00:00:00Z");
  });
});

describe("patchTaskInList", () => {
  const list: TaskListResponse = {
    tasks: [
      makeTask({ id: "t1", title: "One" }),
      makeTask({ id: "t2", title: "Two" }),
    ],
    limit: 20,
    offset: 0,
    has_more: false,
  };

  it("patches the matching row and leaves siblings", () => {
    const next = patchTaskInList(list, "t1", basePatch);
    expect(next?.tasks[0]?.title).toBe("Patched");
    expect(next?.tasks[1]?.title).toBe("Two");
  });

  it("returns null when the task id is absent", () => {
    expect(patchTaskInList(list, "missing", basePatch)).toBeNull();
  });
});
