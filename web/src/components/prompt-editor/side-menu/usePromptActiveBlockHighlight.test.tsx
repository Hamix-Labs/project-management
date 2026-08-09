// @vitest-environment jsdom
import { act, renderHook } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

let editorChange: (() => void) | null = null;

vi.mock("@blocknote/react", () => ({
  useEditorChange: (callback: () => void) => {
    editorChange = callback;
  },
}));

import { PROMPT_BLOCK_ACTIVE_ATTR } from "./promptBlockElement";
import { usePromptActiveBlockHighlight } from "./usePromptActiveBlockHighlight";

function appendBlock(editorDom: HTMLElement, blockId: string) {
  const container = document.createElement("div");
  container.setAttribute("data-node-type", "blockContainer");
  container.setAttribute("data-id", blockId);
  editorDom.appendChild(container);
  return container;
}

describe("usePromptActiveBlockHighlight", () => {
  let editorDom: HTMLElement;

  beforeEach(() => {
    editorChange = null;
    editorDom = document.createElement("div");
    document.body.appendChild(editorDom);
  });

  it("stamps the active attribute on the listed blocks", () => {
    const a = appendBlock(editorDom, "a");
    const b = appendBlock(editorDom, "b");

    renderHook(() => usePromptActiveBlockHighlight(editorDom, ["a", "b"]));

    expect(a.getAttribute(PROMPT_BLOCK_ACTIVE_ATTR)).toBe("true");
    expect(b.getAttribute(PROMPT_BLOCK_ACTIVE_ATTR)).toBe("true");
  });

  it("clears the attribute when ids are removed", () => {
    const a = appendBlock(editorDom, "a");
    const { rerender } = renderHook(
      ({ ids }) => usePromptActiveBlockHighlight(editorDom, ids),
      { initialProps: { ids: ["a"] as string[] } },
    );

    expect(a.getAttribute(PROMPT_BLOCK_ACTIVE_ATTR)).toBe("true");

    rerender({ ids: [] });

    expect(a.hasAttribute(PROMPT_BLOCK_ACTIVE_ATTR)).toBe(false);
  });

  it("re-stamps after ProseMirror replaces the block node", () => {
    appendBlock(editorDom, "a");
    renderHook(() => usePromptActiveBlockHighlight(editorDom, ["a"]));

    editorDom.replaceChildren();
    const replacement = appendBlock(editorDom, "a");
    expect(replacement.hasAttribute(PROMPT_BLOCK_ACTIVE_ATTR)).toBe(false);

    act(() => {
      editorChange?.();
    });

    expect(replacement.getAttribute(PROMPT_BLOCK_ACTIVE_ATTR)).toBe("true");
  });

  it("clears every stamp on unmount", () => {
    const a = appendBlock(editorDom, "a");
    const { unmount } = renderHook(() =>
      usePromptActiveBlockHighlight(editorDom, ["a"]),
    );

    unmount();

    expect(a.hasAttribute(PROMPT_BLOCK_ACTIVE_ATTR)).toBe(false);
  });
});
