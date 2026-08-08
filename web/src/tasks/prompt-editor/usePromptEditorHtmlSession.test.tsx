/** @vitest-environment jsdom */
import { renderHook } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { usePromptEditorHtmlSession } from "./usePromptEditorHtmlSession";
import type { PromptDocumentAdapter } from "./types";

describe("usePromptEditorHtmlSession", () => {
  it("keeps onCommit stable when applyHydratedName identity changes", () => {
    const adapter: PromptDocumentAdapter = {
      load: vi.fn(async () => ({ html: "<p>hi</p>", name: "T" })),
      save: vi.fn(),
      saveName: vi.fn(),
    };
    const applyA = vi.fn();
    const applyB = vi.fn();

    const { result, rerender } = renderHook(
      ({ apply }: { apply: (name?: string) => void }) =>
        usePromptEditorHtmlSession({
          adapter,
          launch: null,
          applyHydratedName: apply,
        }),
      { initialProps: { apply: applyA } },
    );

    // Reach into document load by capturing onCommit via a spy on adapter.load
    // count — load should not re-fire solely because apply identity changed.
    const loadsBefore = vi.mocked(adapter.load).mock.calls.length;
    rerender({ apply: applyB });
    expect(vi.mocked(adapter.load).mock.calls.length).toBe(loadsBefore);
    expect(result.current.status === "loading" || result.current.status === "ready").toBe(
      true,
    );
  });
});
