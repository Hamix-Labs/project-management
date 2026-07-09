import { describe, expect, it } from "vitest";
import { TASK_TEST_DEFAULTS } from "@/test/taskDefaults";
import { flattenTaskTreeRoots } from "./flattenTaskTree";

const task = {
  id: "r1",
  title: "Root",
  initial_prompt: "",
  status: "ready" as const,
  priority: "medium" as const,
  ...TASK_TEST_DEFAULTS,
};

describe("flattenTaskTreeRoots", () => {
  it("returns flat rows at depth 0", () => {
    const flat = flattenTaskTreeRoots([task]);
    expect(flat).toHaveLength(1);
    expect(flat[0]).toMatchObject({ id: "r1", depth: 0 });
  });

  it("returns an empty array for an empty list", () => {
    expect(flattenTaskTreeRoots([])).toEqual([]);
  });

  it("preserves order", () => {
    const a = {
      id: "a",
      title: "A",
      initial_prompt: "",
      status: "ready" as const,
      priority: "low" as const,
      ...TASK_TEST_DEFAULTS,
    };
    const b = {
      id: "b",
      title: "B",
      initial_prompt: "",
      status: "done" as const,
      priority: "high" as const,
      ...TASK_TEST_DEFAULTS,
    };
    const flat = flattenTaskTreeRoots([a, b]);
    expect(flat.map((t) => t.id)).toEqual(["a", "b"]);
    expect(flat.every((t) => t.depth === 0)).toBe(true);
  });
});
