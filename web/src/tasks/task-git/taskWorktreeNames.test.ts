import { describe, expect, it } from "vitest";
import {
  branchPathSlug,
  predictedTaskWorktreeName,
  taskBranchName,
} from "./taskWorktreeNames";

describe("taskWorktreeNames", () => {
  it("matches Go TaskBranchName for a UUID", () => {
    expect(taskBranchName("0acaf529-9adf-4333-8992-29aa308eadba")).toBe(
      "hamix/task-0acaf529",
    );
  });

  it("slugs branch names like Go BranchPathSlug", () => {
    expect(branchPathSlug("hamix/task-0acaf529")).toBe("hamix-task-0acaf529");
  });

  it("predicts the worktree folder name from task id", () => {
    expect(
      predictedTaskWorktreeName("0acaf529-9adf-4333-8992-29aa308eadba"),
    ).toBe("hamix-task-0acaf529");
  });
});
