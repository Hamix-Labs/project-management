import { describe, expect, it } from "vitest";
import { TASK_TEST_DEFAULTS } from "@/test/taskDefaults";
import { flattenTaskTreeRoots, sortWorktreeFamilyTasks } from "./flattenTaskTree";

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

  it("assigns depth 0 to worktree root and depth 1 to siblings when family active", () => {
    const rootId = "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa";
    const root = {
      id: rootId,
      title: "Root",
      initial_prompt: "",
      status: "ready" as const,
      priority: "medium" as const,
      worktree_id: "wt-1",
      worktree_root_task_id: rootId,
      ...TASK_TEST_DEFAULTS,
    };
    const child = {
      id: "bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb",
      title: "Child",
      initial_prompt: "",
      status: "ready" as const,
      priority: "low" as const,
      worktree_id: "wt-1",
      worktree_root_task_id: rootId,
      ...TASK_TEST_DEFAULTS,
    };
    const flat = flattenTaskTreeRoots([root, child], {
      worktreeFamilyActive: true,
    });
    expect(flat.map((t) => ({ id: t.id, depth: t.depth }))).toEqual([
      { id: rootId, depth: 0 },
      { id: child.id, depth: 1 },
    ]);
  });
});

describe("sortWorktreeFamilyTasks", () => {
  it("orders root before children, then by created_at desc", () => {
    const rootId = "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa";
    const root = {
      id: rootId,
      title: "Root",
      initial_prompt: "",
      status: "ready" as const,
      priority: "medium" as const,
      worktree_id: "wt-1",
      worktree_root_task_id: rootId,
      created_at: "2026-01-01T00:00:00Z",
      ...TASK_TEST_DEFAULTS,
    };
    const olderChild = {
      id: "bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb",
      title: "Older child",
      initial_prompt: "",
      status: "ready" as const,
      priority: "low" as const,
      worktree_id: "wt-1",
      worktree_root_task_id: rootId,
      created_at: "2026-01-02T00:00:00Z",
      ...TASK_TEST_DEFAULTS,
    };
    const newerChild = {
      id: "cccccccc-cccc-4ccc-8ccc-cccccccccccc",
      title: "Newer child",
      initial_prompt: "",
      status: "ready" as const,
      priority: "high" as const,
      worktree_id: "wt-1",
      worktree_root_task_id: rootId,
      created_at: "2026-01-03T00:00:00Z",
      ...TASK_TEST_DEFAULTS,
    };
    expect(
      sortWorktreeFamilyTasks([newerChild, olderChild, root]).map((t) => t.id),
    ).toEqual([rootId, newerChild.id, olderChild.id]);
  });
});
