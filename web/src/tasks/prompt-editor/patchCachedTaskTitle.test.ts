import { describe, expect, it } from "vitest";
import { QueryClient } from "@tanstack/react-query";
import { taskQueryKeys } from "@/lib/taskQueryKeys";
import type { Task, TaskListResponse } from "@/types";
import { patchCachedTaskTitle } from "./patchCachedTaskTitle";

function makeTask(partial: Partial<Task> & { id: string; title: string }): Task {
  return {
    initial_prompt: "",
    status: "todo",
    priority: "medium",
    runner: "cursor",
    cursor_model: "",
    ...partial,
  } as Task;
}

describe("patchCachedTaskTitle", () => {
  it("updates detail and list caches", () => {
    const qc = new QueryClient();
    const task = makeTask({ id: "t1", title: "Old" });
    qc.setQueryData(taskQueryKeys.detail("t1"), task);
    const list: TaskListResponse = {
      tasks: [task, makeTask({ id: "t2", title: "Other" })],
      limit: 20,
      offset: 0,
      has_more: false,
    };
    qc.setQueryData(taskQueryKeys.list({ limit: 20, offset: 0 }), list);

    patchCachedTaskTitle(qc, "t1", "New");

    expect(qc.getQueryData<Task>(taskQueryKeys.detail("t1"))?.title).toBe(
      "New",
    );
    const next = qc.getQueryData<TaskListResponse>(
      taskQueryKeys.list({ limit: 20, offset: 0 }),
    );
    expect(next?.tasks[0]?.title).toBe("New");
    expect(next?.tasks[1]?.title).toBe("Other");
  });
});
