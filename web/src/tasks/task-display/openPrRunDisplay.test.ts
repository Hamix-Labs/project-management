import { describe, expect, it } from "vitest";
import {
  CREATING_PR_STATUS_LABEL,
  isOpenPrRunKind,
  openPrSessionClearedByStatus,
  shouldShowCreatingPrLabel,
} from "./openPrRunDisplay";

describe("openPrRunDisplay", () => {
  it("detects open_pr run kind", () => {
    expect(isOpenPrRunKind({ run_kind: "open_pr" })).toBe(true);
    expect(isOpenPrRunKind({ run_kind: "polish" })).toBe(false);
    expect(isOpenPrRunKind({})).toBe(false);
    expect(isOpenPrRunKind(undefined)).toBe(false);
  });

  it("exports Creating PR label", () => {
    expect(CREATING_PR_STATUS_LABEL).toBe("Creating PR");
  });

  it("shows Creating PR while pending, sticky, or cycle meta known", () => {
    expect(
      shouldShowCreatingPrLabel({
        mutationPending: true,
        sessionActive: false,
        hasRunningOpenPrCycle: false,
      }),
    ).toBe(true);
    expect(
      shouldShowCreatingPrLabel({
        mutationPending: false,
        sessionActive: true,
        hasRunningOpenPrCycle: false,
      }),
    ).toBe(true);
    expect(
      shouldShowCreatingPrLabel({
        mutationPending: false,
        sessionActive: false,
        hasRunningOpenPrCycle: true,
      }),
    ).toBe(true);
    expect(
      shouldShowCreatingPrLabel({
        mutationPending: false,
        sessionActive: false,
        hasRunningOpenPrCycle: false,
      }),
    ).toBe(false);
  });

  it("clears sticky session once status leaves ready/running", () => {
    expect(openPrSessionClearedByStatus("ready")).toBe(false);
    expect(openPrSessionClearedByStatus("running")).toBe(false);
    expect(openPrSessionClearedByStatus("pr_ready")).toBe(true);
    expect(openPrSessionClearedByStatus("failed")).toBe(true);
    expect(openPrSessionClearedByStatus("review")).toBe(true);
  });
});
