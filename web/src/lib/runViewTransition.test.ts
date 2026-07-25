import { afterEach, describe, expect, it, vi } from "vitest";
import { runViewTransition } from "./runViewTransition";

describe("runViewTransition", () => {
  afterEach(() => {
    vi.unstubAllGlobals();
    vi.restoreAllMocks();
  });

  it("returns false and does not call update when unsupported", () => {
    const update = vi.fn();
    expect(runViewTransition(update)).toBe(false);
    expect(update).not.toHaveBeenCalled();
  });

  it("returns false under prefers-reduced-motion", () => {
    const startViewTransition = vi.fn();
    vi.stubGlobal("document", {
      ...document,
      startViewTransition,
    });
    vi.spyOn(window, "matchMedia").mockReturnValue({
      matches: true,
      media: "(prefers-reduced-motion: reduce)",
      onchange: null,
      addListener: () => {},
      removeListener: () => {},
      addEventListener: () => {},
      removeEventListener: () => {},
      dispatchEvent: () => false,
    } as MediaQueryList);

    const update = vi.fn();
    expect(runViewTransition(update)).toBe(false);
    expect(startViewTransition).not.toHaveBeenCalled();
    expect(update).not.toHaveBeenCalled();
  });

  it("starts a view transition when supported", () => {
    const startViewTransition = vi.fn((cb: () => void) => {
      cb();
      return {};
    });
    vi.stubGlobal("document", {
      ...document,
      startViewTransition,
    });
    vi.spyOn(window, "matchMedia").mockReturnValue({
      matches: false,
      media: "(prefers-reduced-motion: reduce)",
      onchange: null,
      addListener: () => {},
      removeListener: () => {},
      addEventListener: () => {},
      removeEventListener: () => {},
      dispatchEvent: () => false,
    } as MediaQueryList);

    const update = vi.fn();
    expect(runViewTransition(update)).toBe(true);
    expect(startViewTransition).toHaveBeenCalledOnce();
    expect(update).toHaveBeenCalledOnce();
  });
});
