import { describe, expect, it } from "vitest";
import type { TaskWithDepth } from "../../../task-tree";
import { buildTaskListRowsToRender } from "./taskListRowAnimations";

function task(id: string): TaskWithDepth {
  return {
    id,
    title: id,
    initial_prompt: "",
    status: "ready",
    priority: "medium",
    runner: "cursor",
    cursor_model: "",
    depth: 0,
  };
}

describe("buildTaskListRowsToRender", () => {
  it("includes visible filtered rows in render order", () => {
    const a = task("a");
    const b = task("b");
    const rows = buildTaskListRowsToRender(
      [a, b],
      [a, b],
      new Map([
        ["a", a],
        ["b", b],
      ]),
      new Set(),
      { current: new Map() },
      { current: new Map() },
      new Set(),
    );
    expect(rows.map((r) => r.task.id)).toEqual(["a", "b"]);
    expect(rows.every((r) => !r.isExiting)).toBe(true);
  });

  it("marks filter-exit rows as exiting", () => {
    const a = task("a");
    const rows = buildTaskListRowsToRender(
      [a],
      [],
      new Map(),
      new Set(["a"]),
      { current: new Map([["a", a]]) },
      { current: new Map() },
      new Set(),
    );
    expect(rows).toHaveLength(1);
    expect(rows[0]?.isExiting).toBe(true);
    expect(rows[0]?.isFilterExit).toBe(true);
  });
});
