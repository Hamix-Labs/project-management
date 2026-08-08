import { QueryClient } from "@tanstack/react-query";
import { describe, expect, it, vi } from "vitest";
import { taskQueryKeys } from "@/lib/taskQueryKeys";
import type { TaskInvalidationScope } from "@/lib/queryInvalidation";
import type { PromptSourceKind } from "@/tasks/prompt-editor/types";
import {
  invalidatePromptDocumentCoherence,
  promptDocumentCoherenceScopes,
} from "./invalidatePromptDocumentCoherence";

describe("promptDocumentCoherenceScopes", () => {
  const cases: {
    name: string;
    kind: PromptSourceKind;
    id: string;
    expected: TaskInvalidationScope[];
  }[] = [
    { name: "draft", kind: "draft", id: "d1", expected: [{ scope: "drafts" }] },
    {
      name: "template",
      kind: "template",
      id: "tpl1",
      expected: [{ scope: "templates" }],
    },
    {
      name: "task",
      kind: "task",
      id: "t1",
      expected: [{ scope: "detail", taskId: "t1" }, { scope: "listStats" }],
    },
    { name: "task without id", kind: "task", id: "", expected: [] },
    {
      name: "ephemeral",
      kind: "ephemeral",
      id: "e1",
      expected: [],
    },
  ];

  it.each(cases)("$name maps to the caches that render it", ({
    kind,
    id,
    expected,
  }) => {
    expect(promptDocumentCoherenceScopes(kind, id)).toEqual(expected);
  });
});

describe("invalidatePromptDocumentCoherence", () => {
  it("invalidates the drafts list for a draft document", () => {
    const queryClient = new QueryClient();
    const invalidate = vi.spyOn(queryClient, "invalidateQueries");

    invalidatePromptDocumentCoherence(queryClient, "draft", "d1");

    expect(invalidate).toHaveBeenCalledWith({
      queryKey: taskQueryKeys.drafts(),
    });
  });

  it("leaves the cache alone for ephemeral documents", () => {
    const queryClient = new QueryClient();
    const invalidate = vi.spyOn(queryClient, "invalidateQueries");

    invalidatePromptDocumentCoherence(queryClient, "ephemeral", "e1");

    expect(invalidate).not.toHaveBeenCalled();
  });
});
