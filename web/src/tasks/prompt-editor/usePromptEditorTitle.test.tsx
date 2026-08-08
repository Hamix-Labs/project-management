/** @vitest-environment jsdom */
import { act, renderHook, waitFor } from "@testing-library/react";
import type { ReactNode } from "react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { describe, expect, it, vi } from "vitest";
import { usePromptEditorTitle } from "./usePromptEditorTitle";
import type { PromptDocumentAdapter } from "./types";

function wrap(qc: QueryClient) {
  return function Wrapper({ children }: { children: ReactNode }) {
    return (
      <QueryClientProvider client={qc}>{children}</QueryClientProvider>
    );
  };
}

describe("usePromptEditorTitle", () => {
  it("rejects empty commits and keeps the previous title", async () => {
    const saveName = vi.fn(async () => undefined);
    const adapter: PromptDocumentAdapter = {
      load: vi.fn(),
      save: vi.fn(),
      saveName,
    };
    const setSessionError = vi.fn();
    const qc = new QueryClient();
    const { result } = renderHook(
      () =>
        usePromptEditorTitle({
          launchTitle: "Keep me",
          adapter,
          sourceKind: "draft",
          sourceId: "d1",
          setSessionError,
          onDocumentSaved: vi.fn(),
        }),
      { wrapper: wrap(qc) },
    );

    await act(async () => {
      await result.current.onTitleCommit("   ");
    });
    expect(result.current.title).toBe("Keep me");
    expect(saveName).not.toHaveBeenCalled();
  });

  it("reverts and surfaces error when saveName fails", async () => {
    const adapter: PromptDocumentAdapter = {
      load: vi.fn(),
      save: vi.fn(),
      saveName: vi.fn(async () => {
        throw new Error("rename boom");
      }),
    };
    const setSessionError = vi.fn();
    const qc = new QueryClient();
    const { result } = renderHook(
      () =>
        usePromptEditorTitle({
          launchTitle: "Original",
          adapter,
          sourceKind: "task",
          sourceId: "t1",
          setSessionError,
          onDocumentSaved: vi.fn(),
        }),
      { wrapper: wrap(qc) },
    );

    await act(async () => {
      await result.current.onTitleCommit("Attempt");
    });

    await waitFor(() => expect(result.current.title).toBe("Original"));
    expect(setSessionError).toHaveBeenCalled();
    expect(setSessionError.mock.calls[0][0].code).toBe("rename_failed");
  });

  it("refreshes owning caches after a successful rename", async () => {
    const adapter: PromptDocumentAdapter = {
      load: vi.fn(),
      save: vi.fn(),
      saveName: vi.fn(async () => undefined),
    };
    const onDocumentSaved = vi.fn();
    const qc = new QueryClient();
    const { result } = renderHook(
      () =>
        usePromptEditorTitle({
          launchTitle: "Old name",
          adapter,
          sourceKind: "draft",
          sourceId: "d1",
          setSessionError: vi.fn(),
          onDocumentSaved,
        }),
      { wrapper: wrap(qc) },
    );

    await act(async () => {
      await result.current.onTitleCommit("New name");
    });

    expect(adapter.saveName).toHaveBeenCalledWith("New name");
    expect(onDocumentSaved).toHaveBeenCalledTimes(1);
  });

  it("prefers hydrated name from load", () => {
    const qc = new QueryClient();
    const { result } = renderHook(
      () =>
        usePromptEditorTitle({
          launchTitle: "Launch",
          adapter: null,
          sourceKind: "task",
          sourceId: "t1",
          setSessionError: vi.fn(),
          onDocumentSaved: vi.fn(),
        }),
      { wrapper: wrap(qc) },
    );
    act(() => {
      result.current.applyHydratedName("  From server  ");
    });
    expect(result.current.title).toBe("From server");
  });
});
