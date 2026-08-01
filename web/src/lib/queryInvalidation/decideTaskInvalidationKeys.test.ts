import { describe, expect, it } from "vitest";
import { taskQueryKeys } from "@/lib/taskQueryKeys";
import { decideTaskInvalidationKeys } from "./decideTaskInvalidationKeys";
import type { TaskInvalidationScope } from "./types";

describe("decideTaskInvalidationKeys", () => {
  const cases: {
    name: string;
    input: TaskInvalidationScope;
    expected: readonly (readonly unknown[])[];
  }[] = [
    {
      name: "listStats",
      input: { scope: "listStats" },
      expected: [taskQueryKeys.listRoot(), taskQueryKeys.stats()],
    },
    {
      name: "detail",
      input: { scope: "detail", taskId: "task-1" },
      expected: [taskQueryKeys.detail("task-1")],
    },
    {
      name: "checklist",
      input: { scope: "checklist", taskId: "task-1" },
      expected: [
        taskQueryKeys.checklist("task-1"),
        taskQueryKeys.detail("task-1"),
      ],
    },
    {
      name: "events",
      input: { scope: "events", taskId: "task-1" },
      expected: [taskQueryKeys.eventsRoot("task-1")],
    },
    {
      name: "cycles",
      input: { scope: "cycles", taskId: "task-1" },
      expected: [taskQueryKeys.cycles("task-1")],
    },
    {
      name: "drafts",
      input: { scope: "drafts" },
      expected: [taskQueryKeys.drafts()],
    },
    {
      name: "templates",
      input: { scope: "templates" },
      expected: [taskQueryKeys.templates()],
    },
  ];

  it.each(cases)("$name returns the catalog keys", ({ input, expected }) => {
    expect(decideTaskInvalidationKeys(input)).toEqual(expected);
  });
});
