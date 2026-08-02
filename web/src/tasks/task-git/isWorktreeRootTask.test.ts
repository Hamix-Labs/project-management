import { describe, expect, it } from "vitest";
import { isWorktreeRootTask } from "./isWorktreeRootTask";

describe("isWorktreeRootTask", () => {
  it("returns true when root id matches", () => {
    expect(
      isWorktreeRootTask({ id: "t1", worktree_root_task_id: "t1" }),
    ).toBe(true);
  });

  it("returns false for non-root sibling", () => {
    expect(
      isWorktreeRootTask({ id: "t2", worktree_root_task_id: "t1" }),
    ).toBe(false);
  });

  it("returns true when root is unknown", () => {
    expect(isWorktreeRootTask({ id: "t1" })).toBe(true);
  });
});
