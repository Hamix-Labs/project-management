import { describe, expect, it } from "vitest";
import { taskQueryKeys } from "@/lib/taskQueryKeys";

describe("taskQueryKeys", () => {
  it("builds list root and per-page list keys with limit/offset params", () => {
    expect(taskQueryKeys.listRoot()).toEqual(["tasks", "list"]);
    expect(taskQueryKeys.list({ limit: 20, offset: 0 })).toEqual([
      "tasks",
      "list",
      { limit: 20, offset: 0 },
    ]);
    expect(taskQueryKeys.list({ limit: 200, offset: 0 })).toEqual([
      "tasks",
      "list",
      { limit: 200, offset: 0 },
    ]);
    expect(taskQueryKeys.list({ limit: 20, offset: 60 })).toEqual([
      "tasks",
      "list",
      { limit: 20, offset: 60 },
    ]);
  });

  it("places board key under listRoot for invalidation and optimistic updates", () => {
    expect(taskQueryKeys.board()).toEqual(["tasks", "list", "board"]);
    expect(taskQueryKeys.board().slice(0, 2)).toEqual(taskQueryKeys.listRoot());
    expect(taskQueryKeys.board({ worktreeId: "wt-1" })).toEqual([
      "tasks",
      "list",
      "board",
      { worktreeId: "wt-1" },
    ]);
  });

  it("includes worktreeId in list keys when filtering by family", () => {
    expect(
      taskQueryKeys.list({ limit: 20, offset: 0, worktreeId: "wt-1" }),
    ).toEqual(["tasks", "list", { limit: 20, offset: 0, worktreeId: "wt-1" }]);
  });

  it("defines eventsRoot prefix covering paged and infinite event queries", () => {
    const root = taskQueryKeys.eventsRoot("t1");
    expect(root).toEqual(["tasks", "detail", "t1", "events"]);
    expect(taskQueryKeys.events("t1", { k: "head" }).slice(0, root.length)).toEqual(
      root,
    );
    expect(taskQueryKeys.eventsInfinite("t1").slice(0, root.length)).toEqual(root);
  });

  it("scopes detail, checklist, and event detail under the task id", () => {
    expect(taskQueryKeys.detailRoot()).toEqual(["tasks", "detail"]);
    expect(taskQueryKeys.detail("t1")).toEqual(["tasks", "detail", "t1"]);
    expect(taskQueryKeys.checklist("t1")).toEqual([
      "tasks",
      "detail",
      "t1",
      "checklist",
    ]);
    expect(taskQueryKeys.eventDetail("t1", 42)).toEqual([
      "tasks",
      "detail",
      "t1",
      "event",
      42,
    ]);
  });

  it("scopes cycles list and per-cycle keys under the task detail", () => {
    expect(taskQueryKeys.cycles("t1")).toEqual([
      "tasks",
      "detail",
      "t1",
      "cycles",
    ]);
    expect(taskQueryKeys.cycle("t1", "cyc-1")).toEqual([
      "tasks",
      "detail",
      "t1",
      "cycles",
      "cyc-1",
    ]);
    expect(taskQueryKeys.tokenUsage("t1")).toEqual([
      "tasks",
      "detail",
      "t1",
      "token-usage",
    ]);
  });

  it("encodes events cursor variants in the key", () => {
    expect(taskQueryKeys.events("t1", { k: "head" })).toEqual([
      "tasks",
      "detail",
      "t1",
      "events",
      "head",
    ]);
    expect(taskQueryKeys.events("t1", { k: "before", seq: 9 })).toEqual([
      "tasks",
      "detail",
      "t1",
      "events",
      "before",
      9,
    ]);
    expect(taskQueryKeys.events("t1", { k: "after", seq: 10 })).toEqual([
      "tasks",
      "detail",
      "t1",
      "events",
      "after",
      10,
    ]);
  });

  it("defines infinite events key for bidirectional task detail feed", () => {
    expect(taskQueryKeys.eventsInfinite("t1")).toEqual([
      "tasks",
      "detail",
      "t1",
      "events",
      "infinite",
    ]);
  });

  it("defines stats, drafts, and templates keys for invalidation outside tasks tree", () => {
    expect(taskQueryKeys.stats()).toEqual(["task-stats"]);
    expect(taskQueryKeys.drafts()).toEqual(["task-drafts"]);
    expect(taskQueryKeys.templates()).toEqual(["task-templates"]);
    expect(taskQueryKeys.templates({ q: "alpha" })).toEqual([
      "task-templates",
      { q: "alpha" },
    ]);
  });

  it("scopes cycle failures list queries for pagination and sort", () => {
    expect(taskQueryKeys.cycleFailuresRoot()).toEqual([
      "tasks",
      "cycle-failures",
    ]);
    expect(taskQueryKeys.cycleFailures("at_desc", 0)).toEqual([
      "tasks",
      "cycle-failures",
      "at_desc",
      0,
    ]);
  });

  it("scopes activity feed queries under tasks/activity", () => {
    expect(taskQueryKeys.activityRoot()).toEqual(["tasks", "activity"]);
    expect(taskQueryKeys.activity(undefined, 0)).toEqual([
      "tasks",
      "activity",
      "",
      0,
    ]);
    expect(taskQueryKeys.activity("2026-07-18T00:00:00.000Z", 0)).toEqual([
      "tasks",
      "activity",
      "2026-07-18T00:00:00.000Z",
      0,
    ]);
  });
});
