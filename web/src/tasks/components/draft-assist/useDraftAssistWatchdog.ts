import { useEffect, useMemo, useRef, useState } from "react";

/**
 * Watchdog thresholds (see `docs/design/task-draft-ai.md` §Silence watchdog).
 *
 * - 8s: `Still working (Ns)` — plain reassurance.
 * - 30s: same copy plus a Cancel affordance.
 * - 90s: `This is taking too long. Retry.` — do not fail silently.
 *
 * Kept as constants so tests can assert phase transitions without
 * hard-coding magic numbers.
 */
export const DRAFT_ASSIST_WATCHDOG_STILL_WORKING_MS = 8_000;
export const DRAFT_ASSIST_WATCHDOG_OFFER_CANCEL_MS = 30_000;
export const DRAFT_ASSIST_WATCHDOG_TOO_LONG_MS = 90_000;

export type DraftAssistWatchdogPhase =
  | "none"
  | "still_working"
  | "offer_cancel"
  | "too_long";

export type DraftAssistWatchdogResult = {
  /** Milliseconds since `startedAt` (0 when inactive). */
  elapsedMs: number;
  phase: DraftAssistWatchdogPhase;
  /** Human copy that hangs below the primary status pill. */
  message: string;
  /** True from the 30s threshold — UI shows Cancel/Stop even if the run seems wedged. */
  offerCancel: boolean;
  /** True from the 90s threshold — UI shows Retry and stops updating elapsed. */
  offerRetry: boolean;
};

const INACTIVE: DraftAssistWatchdogResult = {
  elapsedMs: 0,
  phase: "none",
  message: "",
  offerCancel: false,
  offerRetry: false,
};

/**
 * Tick a 1 Hz clock while `active` is true and derive the watchdog
 * phase from the elapsed time since `startedAt`. Callers pass the
 * timestamp from the run-start action so a reconnect does not reset
 * the "Still working" copy.
 *
 * The hook uses `Date.now()` (or `nowFn`) rather than `performance.now()`
 * so tests can stub via `vi.useFakeTimers()` and stay deterministic.
 */
export function useDraftAssistWatchdog(
  active: boolean,
  startedAt: number | null,
  options?: { nowFn?: () => number; tickMs?: number },
): DraftAssistWatchdogResult {
  const nowFn = options?.nowFn ?? Date.now;
  const tickMs = options?.tickMs ?? 1_000;
  const [now, setNow] = useState<number>(() => nowFn());
  const nowFnRef = useRef(nowFn);
  useEffect(() => {
    nowFnRef.current = nowFn;
  }, [nowFn]);

  useEffect(() => {
    if (!active || startedAt == null) return;
    setNow(nowFnRef.current());
    const t = setInterval(() => {
      setNow(nowFnRef.current());
    }, tickMs);
    return () => clearInterval(t);
  }, [active, startedAt, tickMs]);

  return useMemo<DraftAssistWatchdogResult>(() => {
    if (!active || startedAt == null) return INACTIVE;
    const elapsedMs = Math.max(0, now - startedAt);
    const phase = phaseForElapsed(elapsedMs);
    const seconds = Math.floor(elapsedMs / 1_000);
    return {
      elapsedMs,
      phase,
      message: messageForPhase(phase, seconds),
      offerCancel:
        phase === "offer_cancel" || phase === "too_long",
      offerRetry: phase === "too_long",
    };
  }, [active, startedAt, now]);
}

/** Pure — exported so tests can pin the phase math without a React tree. */
export function phaseForElapsed(elapsedMs: number): DraftAssistWatchdogPhase {
  if (elapsedMs >= DRAFT_ASSIST_WATCHDOG_TOO_LONG_MS) return "too_long";
  if (elapsedMs >= DRAFT_ASSIST_WATCHDOG_OFFER_CANCEL_MS) return "offer_cancel";
  if (elapsedMs >= DRAFT_ASSIST_WATCHDOG_STILL_WORKING_MS) return "still_working";
  return "none";
}

function messageForPhase(
  phase: DraftAssistWatchdogPhase,
  seconds: number,
): string {
  switch (phase) {
    case "none":
      return "";
    case "still_working":
    case "offer_cancel":
      return `Still working (${seconds}s)`;
    case "too_long":
      return "This is taking too long.";
  }
}
