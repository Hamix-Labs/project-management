import { QueryClient } from "@tanstack/react-query";
import { describe, expect, it } from "vitest";
import { taskQueryKeys } from "@/lib/taskQueryKeys";
import { makeTask } from "@/test/taskDefaults";
import type { TaskListResponse } from "@/types";
import {
  applyCreatedTaskToCache,
  insertTaskInList,
  patchTaskPickupInList,
  removeTaskFromList,
} from "./optimisticTaskList";

const listPage: TaskListResponse = {
  tasks: [makeTask({ id: "existing", title: "Existing" })],
  limit: 20,
  offset: 0,
  has_more: false,
};

describe("optimisticTaskList", () => {
  it("insertTaskInList prepends when the task is new", () => {
    const created = makeTask({ id: "new-1", title: "New" });
    const next = insertTaskInList(listPage, created);
    expect(next?.tasks.map((t) => t.id)).toEqual(["new-1", "existing"]);
  });

  it("removeTaskFromList drops a row by id", () => {
    const next = removeTaskFromList(listPage, "existing");
    expect(next?.tasks).toHaveLength(0);
  });

  it("patchTaskPickupInList updates schedule on one row", () => {
    const next = patchTaskPickupInList(listPage, "existing", "2026-07-08T12:00:00Z");
    expect(next?.tasks[0]?.pickup_not_before).toBe("2026-07-08T12:00:00Z");
  });

  it("applyCreatedTaskToCache seeds detail and list caches", () => {
    const queryClient = new QueryClient();
    const listKey = taskQueryKeys.list({ limit: 20, offset: 0 });
    queryClient.setQueryData(listKey, listPage);
    const created = makeTask({ id: "created-1", title: "Created" });

    applyCreatedTaskToCache(queryClient, created);

    expect(queryClient.getQueryData(taskQueryKeys.detail("created-1"))).toEqual(
      created,
    );
    const list = queryClient.getQueryData<TaskListResponse>(listKey);
    expect(list?.tasks.map((t) => t.id)).toEqual(["created-1", "existing"]);
  });
});
