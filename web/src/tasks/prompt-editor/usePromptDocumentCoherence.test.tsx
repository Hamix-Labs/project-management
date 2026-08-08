/** @vitest-environment jsdom */
import { act, renderHook } from "@testing-library/react";
import type { ReactNode } from "react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { describe, expect, it, vi } from "vitest";
import { taskQueryKeys } from "@/lib/taskQueryKeys";
import { usePromptDocumentCoherence } from "./usePromptDocumentCoherence";

function wrap(qc: QueryClient) {
  return function Wrapper({ children }: { children: ReactNode }) {
    return <QueryClientProvider client={qc}>{children}</QueryClientProvider>;
  };
}

describe("usePromptDocumentCoherence", () => {
  it("refreshes the drafts list when the editor unmounts", () => {
    const qc = new QueryClient();
    const invalidate = vi.spyOn(qc, "invalidateQueries");
    const { unmount } = renderHook(
      () => usePromptDocumentCoherence("draft", "d1"),
      { wrapper: wrap(qc) },
    );

    expect(invalidate).not.toHaveBeenCalled();

    unmount();

    expect(invalidate).toHaveBeenCalledWith({
      queryKey: taskQueryKeys.drafts(),
    });
  });

  it("refreshes on demand for writes the user must see right away", () => {
    const qc = new QueryClient();
    const invalidate = vi.spyOn(qc, "invalidateQueries");
    const { result } = renderHook(
      () => usePromptDocumentCoherence("draft", "d1"),
      { wrapper: wrap(qc) },
    );

    act(() => result.current());

    expect(invalidate).toHaveBeenCalledWith({
      queryKey: taskQueryKeys.drafts(),
    });
  });

  it("stays a no-op for ephemeral documents and unknown kinds", () => {
    const qc = new QueryClient();
    const invalidate = vi.spyOn(qc, "invalidateQueries");
    const ephemeral = renderHook(
      () => usePromptDocumentCoherence("ephemeral", "e1"),
      { wrapper: wrap(qc) },
    );
    const unknown = renderHook(
      () => usePromptDocumentCoherence("nonsense", "x"),
      { wrapper: wrap(qc) },
    );

    act(() => ephemeral.result.current());
    act(() => unknown.result.current());
    ephemeral.unmount();
    unknown.unmount();

    expect(invalidate).not.toHaveBeenCalled();
  });
});
