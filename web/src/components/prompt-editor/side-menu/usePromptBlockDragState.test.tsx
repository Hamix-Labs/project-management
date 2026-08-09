import { act, renderHook } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
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

function dispatchMouseMove(target: EventTarget) {
  act(() => {
    target.dispatchEvent(new Event("mousemove", { bubbles: true }));
  });
}

/** Runs the task the hook queues from `dragstart`. */
function flushDeferredDragStart() {
  act(() => {
    vi.advanceTimersByTime(1);
  });
}

beforeEach(() => {
  vi.useFakeTimers();
});

afterEach(() => {
  vi.useRealTimers();
  host?.remove();
  host = null;
});

describe("usePromptBlockDragState", () => {
  it("does not flip during the dragstart handler itself", () => {
    const { host: editorHost, inner } = mountHost();
    const { result } = renderHook(() => usePromptBlockDragState(editorHost));

    dispatchDrag(inner, "dragstart");

    // Restyling the drag source from inside dragstart makes Chrome abandon the
    // drag, and the drag handle lives inside the menu this flag hides.
    expect(result.current).toBe(false);
  });

  it("is active between a drag starting in the editor and its drop", () => {
    const { host: editorHost, inner } = mountHost();
    const { result } = renderHook(() => usePromptBlockDragState(editorHost));

    expect(result.current).toBe(false);

    dispatchDrag(inner, "dragstart");
    flushDeferredDragStart();
    expect(result.current).toBe(true);

    dispatchDrag(inner, "drop");
    expect(result.current).toBe(false);
  });

  it("clears on dragend, which is the only event a cancelled drag fires", () => {
    const { host: editorHost, inner } = mountHost();
    const { result } = renderHook(() => usePromptBlockDragState(editorHost));

    dispatchDrag(inner, "dragstart");
    flushDeferredDragStart();

    dispatchDrag(document.body, "dragend");
    expect(result.current).toBe(false);
  });

  it("clears on mousemove, which cannot fire while a drag is in flight", () => {
    const { host: editorHost, inner } = mountHost();
    const { result } = renderHook(() => usePromptBlockDragState(editorHost));

    dispatchDrag(inner, "dragstart");
    flushDeferredDragStart();

    dispatchMouseMove(document.body);
    expect(result.current).toBe(false);
  });

  it("drops a pending dragstart when the drag is cancelled immediately", () => {
    const { host: editorHost, inner } = mountHost();
    const { result } = renderHook(() => usePromptBlockDragState(editorHost));

    dispatchDrag(inner, "dragstart");
    dispatchDrag(document.body, "dragend");
    flushDeferredDragStart();

    expect(result.current).toBe(false);
  });

  it("recognises a BlockNote payload dragged from outside the host", () => {
    const { host: editorHost } = mountHost();
    const outside = document.createElement("div");
    document.body.appendChild(outside);
    const { result } = renderHook(() => usePromptBlockDragState(editorHost));

    try {
      dispatchDrag(outside, "dragstart", ["blocknote/html"]);
      flushDeferredDragStart();
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
      flushDeferredDragStart();
      expect(result.current).toBe(false);
    } finally {
      outside.remove();
    }
  });

  it("stays inert until the editor host exists", () => {
    const { inner } = mountHost();
    const { result } = renderHook(() => usePromptBlockDragState(null));

    dispatchDrag(inner, "dragstart");
    flushDeferredDragStart();

    expect(result.current).toBe(false);
  });
});
