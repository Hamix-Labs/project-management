import { describe, expect, it } from "vitest";
import { STATUS_META, statusListLabel } from "./taskStatusDisplay";

describe("taskStatusDisplay", () => {
  it("capitalizes PR Ready", () => {
    expect(statusListLabel("pr_ready")).toBe("PR Ready");
    expect(STATUS_META.pr_ready.label).toBe("PR Ready");
  });

  it("maps pr_ready to its own tone", () => {
    expect(STATUS_META.pr_ready.tone).toBe("pr_ready");
    expect(STATUS_META.review.tone).toBe("review");
  });
});
