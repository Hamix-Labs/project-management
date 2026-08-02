import { describe, expect, it } from "vitest";
import { TASK_TEST_DEFAULTS } from "@/test/taskDefaults";
import { worktreeFamilyFilterOptions } from "./worktreeFamilyFilterOptions";

describe("worktreeFamilyFilterOptions", () => {
  it("returns unique worktrees labeled by the root task when present", () => {
    const root = {
      id: "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa",
      title: "Root task",
      initial_prompt: "",
      status: "ready" as const,
      priority: "medium" as const,
      worktree_id: "wt-1",
      worktree_root_task_id: "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa",
      ...TASK_TEST_DEFAULTS,
    };
    const child = {
      id: "bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb",
      title: "Child task",
      initial_prompt: "",
      status: "ready" as const,
      priority: "low" as const,
      worktree_id: "wt-1",
      worktree_root_task_id: "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa",
      ...TASK_TEST_DEFAULTS,
    };
    const other = {
      id: "cccccccc-cccc-4ccc-8ccc-cccccccccccc",
      title: "Other family",
      initial_prompt: "",
      status: "ready" as const,
      priority: "high" as const,
      worktree_id: "wt-2",
      worktree_root_task_id: "cccccccc-cccc-4ccc-8ccc-cccccccccccc",
      ...TASK_TEST_DEFAULTS,
    };

    const options = worktreeFamilyFilterOptions([child, other, root]);
    expect(options).toHaveLength(2);
    expect(options.find((o) => o.value === "wt-1")?.label).toContain("Root task");
    expect(options.find((o) => o.value === "wt-2")?.label).toContain("Other family");
  });

  it("skips tasks without a worktree", () => {
    expect(
      worktreeFamilyFilterOptions([
        {
          id: "r1",
          title: "No wt",
          initial_prompt: "",
          status: "ready",
          priority: "medium",
          ...TASK_TEST_DEFAULTS,
        },
      ]),
    ).toEqual([]);
  });
});
