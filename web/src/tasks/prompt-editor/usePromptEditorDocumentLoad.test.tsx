import { act, renderHook, waitFor } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { usePromptEditorDocumentLoad } from "./usePromptEditorDocumentLoad";
import type { PromptDocumentAdapter } from "./types";

describe("usePromptEditorDocumentLoad", () => {
  it("commits snapshot and reaches ready", async () => {
    const adapter: PromptDocumentAdapter = {
      load: vi.fn(async () => ({ html: "<p>hello there friend</p>" })),
      save: vi.fn(),
    };
    const onCommit = vi.fn();
    const onLoadError = vi.fn();
    const onStatus = vi.fn();
    const dirtyRef = { current: false };

    renderHook(() =>
      usePromptEditorDocumentLoad({
        adapter,
        launch: null,
        loadNonce: 0,
        dirtyRef,
        onCommit,
        onLoadError,
        onStatus,
      }),
    );

    await waitFor(() => expect(onStatus).toHaveBeenCalledWith("ready"));
    expect(onCommit).toHaveBeenCalledWith({
      html: "<p>hello there friend</p>",
      worktreeId: undefined,
    });
    expect(onLoadError).not.toHaveBeenCalled();
  });

  it("prefers non-empty seedHtml over adapter html", async () => {
    const adapter: PromptDocumentAdapter = {
      load: vi.fn(async () => ({ html: "<p>server</p>" })),
      save: vi.fn(),
    };
    const onCommit = vi.fn();

    renderHook(() =>
      usePromptEditorDocumentLoad({
        adapter,
        launch: { seedHtml: "<p>seeded content here</p>" },
        loadNonce: 0,
        dirtyRef: { current: false },
        onCommit,
        onLoadError: vi.fn(),
        onStatus: vi.fn(),
      }),
    );

    await waitFor(() =>
      expect(onCommit).toHaveBeenCalledWith({
        html: "<p>seeded content here</p>",
        worktreeId: undefined,
      }),
    );
  });

  it("settles on error when load fails", async () => {
    const adapter: PromptDocumentAdapter = {
      load: vi.fn(async () => {
        throw new Error("draft unavailable");
      }),
      save: vi.fn(),
    };
    const onLoadError = vi.fn();
    const onStatus = vi.fn();

    renderHook(() =>
      usePromptEditorDocumentLoad({
        adapter,
        launch: null,
        loadNonce: 0,
        dirtyRef: { current: false },
        onCommit: vi.fn(),
        onLoadError,
        onStatus,
      }),
    );

    await waitFor(() => expect(onStatus).toHaveBeenCalledWith("error"));
    expect(onLoadError).toHaveBeenCalled();
    expect(onLoadError.mock.calls[0][0].phase).toBe("load");
    expect(onLoadError.mock.calls[0][0].detail).toContain("draft unavailable");
  });

  it("ignores stale load after remount and settles on the latest", async () => {
    let resolveFirst!: (v: { html: string }) => void;
    const first = new Promise<{ html: string }>((r) => {
      resolveFirst = r;
    });
    const load = vi
      .fn()
      .mockImplementationOnce(() => first)
      .mockImplementationOnce(async () => ({ html: "<p>second</p>" }));

    const adapter: PromptDocumentAdapter = { load, save: vi.fn() };
    const onCommit = vi.fn();
    const onStatus = vi.fn();

    const { rerender, unmount } = renderHook(
      ({ nonce }: { nonce: number }) =>
        usePromptEditorDocumentLoad({
          adapter,
          launch: null,
          loadNonce: nonce,
          dirtyRef: { current: false },
          onCommit,
          onLoadError: vi.fn(),
          onStatus,
        }),
      { initialProps: { nonce: 0 } },
    );

    await act(async () => {
      rerender({ nonce: 1 });
    });

    await waitFor(() => expect(onCommit).toHaveBeenCalledWith({
      html: "<p>second</p>",
      worktreeId: undefined,
    }));

    await act(async () => {
      resolveFirst({ html: "<p>first-stale</p>" });
    });

    expect(onCommit).toHaveBeenCalledTimes(1);
    expect(onStatus).toHaveBeenCalledWith("ready");
    unmount();
  });
});
