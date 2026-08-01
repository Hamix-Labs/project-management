import { describe, expect, it } from "vitest";
import { CREATING_PR_STATUS_LABEL, isOpenPrRunKind } from "./openPrRunDisplay";

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
});
