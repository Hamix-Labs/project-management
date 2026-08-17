import { vi } from "vitest";

function emptyRectList(): DOMRectList {
  const list: DOMRect[] = [];
  return {
    item: (index: number) => list[index] ?? null,
    length: 0,
    *[Symbol.iterator]() {},
  } as DOMRectList;
}

type RectHost = {
  getClientRects: () => DOMRectList;
  getBoundingClientRect: () => DOMRect;
};

/** jsdom lacks layout APIs that TipTap / ProseMirror call on mount. */
export function installEditorDomStubs() {
  if (typeof window.matchMedia !== "function") {
    Object.defineProperty(window, "matchMedia", {
      writable: true,
      configurable: true,
      value: (query: string) => ({
        matches: false,
        media: query,
        onchange: null,
        addListener: vi.fn(),
        removeListener: vi.fn(),
        addEventListener: vi.fn(),
        removeEventListener: vi.fn(),
        dispatchEvent: vi.fn(),
      }),
    });
  }

  if (typeof window.ResizeObserver !== "function") {
    class ResizeObserverStub {
      observe() {}
      unobserve() {}
      disconnect() {}
    }
    Object.defineProperty(window, "ResizeObserver", {
      writable: true,
      configurable: true,
      value: ResizeObserverStub,
    });
  }

  const rect = () => new DOMRect(0, 0, 0, 0);
  const hosts: RectHost[] = [
    Element.prototype as unknown as RectHost,
    Range.prototype as unknown as RectHost,
    Text.prototype as unknown as RectHost,
  ];
  for (const host of hosts) {
    host.getClientRects = () => emptyRectList();
    host.getBoundingClientRect = rect;
  }
}
