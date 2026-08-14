// @vitest-environment jsdom
import { renderHook, act } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import {
  DRAFT_ASSIST_WATCHDOG_OFFER_CANCEL_MS,
  DRAFT_ASSIST_WATCHDOG_STILL_WORKING_MS,
  DRAFT_ASSIST_WATCHDOG_TOO_LONG_MS,
  phaseForElapsed,
  useDraftAssistWatchdog,
} from "./useDraftAssistWatchdog";

describe("phaseForElapsed", () => {
  it("maps time to design thresholds", () => {
    expect(phaseForElapsed(0)).toBe("none");
    expect(phaseForElapsed(DRAFT_ASSIST_WATCHDOG_STILL_WORKING_MS - 1)).toBe("none");
    expect(phaseForElapsed(DRAFT_ASSIST_WATCHDOG_STILL_WORKING_MS)).toBe("still_working");
    expect(phaseForElapsed(DRAFT_ASSIST_WATCHDOG_OFFER_CANCEL_MS)).toBe("offer_cancel");
    expect(phaseForElapsed(DRAFT_ASSIST_WATCHDOG_TOO_LONG_MS)).toBe("too_long");
  });
});

describe("useDraftAssistWatchdog", () => {
  let currentMs = 0;
  const nowFn = () => currentMs;

  beforeEach(() => {
    currentMs = 0;
    vi.useFakeTimers();
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it("stays inactive until active + startedAt are set", () => {
    const { result } = renderHook(() => useDraftAssistWatchdog(false, null, { nowFn }));
    expect(result.current.phase).toBe("none");
    expect(result.current.message).toBe("");
  });

  it("crosses 8s → Still working with seconds count", () => {
    currentMs = 0;
    const { result, rerender } = renderHook(
      ({ active, startedAt }: { active: boolean; startedAt: number | null }) =>
        useDraftAssistWatchdog(active, startedAt, { nowFn }),
      { initialProps: { active: true, startedAt: 0 } },
    );
    // Nothing to display yet.
    expect(result.current.phase).toBe("none");

    act(() => {
      currentMs = DRAFT_ASSIST_WATCHDOG_STILL_WORKING_MS + 1_500;
      vi.advanceTimersByTime(DRAFT_ASSIST_WATCHDOG_STILL_WORKING_MS + 1_500);
    });
    rerender({ active: true, startedAt: 0 });
    expect(result.current.phase).toBe("still_working");
    expect(result.current.message).toMatch(/Still working \(9s\)/);
    expect(result.current.offerCancel).toBe(false);
    expect(result.current.offerRetry).toBe(false);
  });

  it("crosses 30s → offer_cancel; 90s → too_long with Retry", () => {
    currentMs = 0;
    const { result, rerender } = renderHook(
      ({ active, startedAt }: { active: boolean; startedAt: number | null }) =>
        useDraftAssistWatchdog(active, startedAt, { nowFn }),
      { initialProps: { active: true, startedAt: 0 } },
    );

    act(() => {
      currentMs = DRAFT_ASSIST_WATCHDOG_OFFER_CANCEL_MS + 500;
      vi.advanceTimersByTime(DRAFT_ASSIST_WATCHDOG_OFFER_CANCEL_MS + 500);
    });
    rerender({ active: true, startedAt: 0 });
    expect(result.current.phase).toBe("offer_cancel");
    expect(result.current.offerCancel).toBe(true);
    expect(result.current.offerRetry).toBe(false);

    act(() => {
      currentMs = DRAFT_ASSIST_WATCHDOG_TOO_LONG_MS + 100;
      vi.advanceTimersByTime(DRAFT_ASSIST_WATCHDOG_TOO_LONG_MS + 100);
    });
    rerender({ active: true, startedAt: 0 });
    expect(result.current.phase).toBe("too_long");
    expect(result.current.offerRetry).toBe(true);
    expect(result.current.message).toBe("This is taking too long.");
  });

  it("resets to inactive when active flips off", () => {
    type Args = { active: boolean; startedAt: number | null };
    const { result, rerender } = renderHook(
      ({ active, startedAt }: Args) =>
        useDraftAssistWatchdog(active, startedAt, { nowFn }),
      { initialProps: { active: true, startedAt: 0 } as Args },
    );
    act(() => {
      currentMs = DRAFT_ASSIST_WATCHDOG_STILL_WORKING_MS + 500;
      vi.advanceTimersByTime(DRAFT_ASSIST_WATCHDOG_STILL_WORKING_MS + 500);
    });
    rerender({ active: true, startedAt: 0 });
    expect(result.current.phase).toBe("still_working");

    rerender({ active: false, startedAt: null });
    expect(result.current.phase).toBe("none");
    expect(result.current.message).toBe("");
  });
});
