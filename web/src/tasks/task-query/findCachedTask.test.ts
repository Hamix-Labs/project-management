import { describe, expect, it } from "vitest";
import { QueryClient } from "@tanstack/react-query";
import { taskQueryKeys } from "@/lib/taskQueryKeys";
import { makeTask } from "@/test/taskDefaults";
import { findCachedTask } from "./findCachedTask";

describe("findCachedTask", () => {
  it("returns detail cache hit", () => {
    const qc = new QueryClient();
    const task = makeTask({ id: "t1", title: "From detail" });
    qc.setQueryData(taskQueryKeys.detail("t1"), task);
    expect(findCachedTask(qc, "t1")?.title).toBe("From detail");
  });

  it("returns list cache hit under listRoot", () => {
    const qc = new QueryClient();
    const task = makeTask({ id: "t2", title: "From list" });
    qc.setQueryData(taskQueryKeys.list({ limit: 20, offset: 0 }), {
      tasks: [task],
      limit: 20,
      offset: 0,
      has_more: false,
    });
    expect(findCachedTask(qc, "t2")?.title).toBe("From list");
  });

  it("returns board cache hit", () => {
    const qc = new QueryClient();
    const task = makeTask({ id: "t3", title: "From board" });
    qc.setQueryData(taskQueryKeys.board(), {
      tasks: [task],
      limit: 50,
      offset: 0,
      has_more: false,
    });
    expect(findCachedTask(qc, "t3")?.title).toBe("From board");
  });

  it("returns undefined when missing", () => {
    const qc = new QueryClient();
    expect(findCachedTask(qc, "missing")).toBeUndefined();
  });
});
