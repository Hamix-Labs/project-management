/** @vitest-environment jsdom */
import { act, renderHook } from "@testing-library/react";
import type { ReactNode } from "react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { describe, expect, it, vi } from "vitest";
import { taskQueryKeys } from "@/lib/taskQueryKeys";
import {
  promptDocumentInvalidationScopes,
  usePromptDocumentInvalidate,
} from "./usePromptDocumentInvalidate";

function wrap(qc: QueryClient) {
  return function Wrapper({ children }: { children: ReactNode }) {
    return <QueryClientProvider client={qc}>{children}</QueryClientProvider>;
  };
}

describe("promptDocumentInvalidationScopes", () => {
  it("maps each source kind to the caches that render it", () => {
    expect(promptDocumentInvalidationScopes("draft", "d1")).toEqual([
      { scope: "drafts" },
    ]);
    expect(promptDocumentInvalidationScopes("template", "t1")).toEqual([
      { scope: "templates" },
    ]);
    expect(promptDocumentInvalidationScopes("task", "t1")).toEqual([
      { scope: "detail", taskId: "t1" },
      { scope: "listStats" },
    ]);
    expect(promptDocumentInvalidationScopes("ephemeral", "e1")).toEqual([]);
  });
});

describe("usePromptDocumentInvalidate", () => {
  it("invalidates the drafts list so a rename shows without a reload", () => {
    const qc = new QueryClient();
    const invalidate = vi.spyOn(qc, "invalidateQueries");
    const { result } = renderHook(
      () => usePromptDocumentInvalidate("draft", "d1"),
      { wrapper: wrap(qc) },
    );

    act(() => result.current());

    expect(invalidate).toHaveBeenCalledWith({
      queryKey: taskQueryKeys.drafts(),
    });
  });

  it("is a no-op for unknown kinds and ephemeral documents", () => {
    const qc = new QueryClient();
    const invalidate = vi.spyOn(qc, "invalidateQueries");
    const { result, rerender } = renderHook(
      ({ kind }: { kind: string }) => usePromptDocumentInvalidate(kind, "x"),
      { wrapper: wrap(qc), initialProps: { kind: "nonsense" } },
    );

    act(() => result.current());
    rerender({ kind: "ephemeral" });
    act(() => result.current());

    expect(invalidate).not.toHaveBeenCalled();
  });
});
