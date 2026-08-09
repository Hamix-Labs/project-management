import { act, renderHook } from "@testing-library/react";
import { afterEach, describe, expect, it } from "vitest";
import { usePromptBlockDragState } from "./usePromptBlockDragState";

let host: HTMLElement | null = null;

function mountHost() {
  host = document.createElement("div");
  const inner = document.createElement("div");
  host.appendChild(inner);
  document.body.appendChild(host);
  return { host, inner };
}

function dispatchDrag(
  target: EventTarget,
  type: "dragstart" | "drop" | "dragend",
  dataTransferTypes?: string[],
) {
  const event = new Event(type, { bubbles: true });
  if (dataTransferTypes) {
    Object.defineProperty(event, "dataTransfer", {
      value: { types: dataTransferTypes },
    });
  }
  act(() => {
    target.dispatchEvent(event);
  });
}

afterEach(() => {
  host?.remove();
  host = null;
});

describe("usePromptBlockDragState", () => {
  it("is active between a drag starting in the editor and its drop", () => {
    const { host: editorHost, inner } = mountHost();
    const { result } = renderHook(() => usePromptBlockDragState(editorHost));

    expect(result.current).toBe(false);

    dispatchDrag(inner, "dragstart");
    expect(result.current).toBe(true);

    dispatchDrag(inner, "drop");
    expect(result.current).toBe(false);
  });

  it("clears on dragend, which is the only event a cancelled drag fires", () => {
    const { host: editorHost, inner } = mountHost();
    const { result } = renderHook(() => usePromptBlockDragState(editorHost));

    dispatchDrag(inner, "dragstart");
    expect(result.current).toBe(true);

    dispatchDrag(document.body, "dragend");
    expect(result.current).toBe(false);
  });

  it("recognises a BlockNote payload dragged from outside the host", () => {
    const { host: editorHost } = mountHost();
    const outside = document.createElement("div");
    document.body.appendChild(outside);
    const { result } = renderHook(() => usePromptBlockDragState(editorHost));

    try {
      dispatchDrag(outside, "dragstart", ["blocknote/html"]);
      expect(result.current).toBe(true);
    } finally {
      outside.remove();
    }
  });

  it("ignores unrelated drags from outside the editor", () => {
    const { host: editorHost } = mountHost();
    const outside = document.createElement("div");
    document.body.appendChild(outside);
    const { result } = renderHook(() => usePromptBlockDragState(editorHost));

    try {
      dispatchDrag(outside, "dragstart", ["Files"]);
      expect(result.current).toBe(false);
    } finally {
      outside.remove();
    }
  });

  it("stays inert until the editor host exists", () => {
    const { inner } = mountHost();
    const { result } = renderHook(() => usePromptBlockDragState(null));

    dispatchDrag(inner, "dragstart");
    expect(result.current).toBe(false);
  });
});
