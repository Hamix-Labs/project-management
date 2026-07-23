import { describe, expect, it } from "vitest";
import { formatShareOfTaskPct, formatTokenCount } from "./formatTokenCount";

describe("formatTokenCount", () => {
  it("formats sub-thousand counts verbatim", () => {
    expect(formatTokenCount(820)).toEqual({
      label: "820",
      ariaLabel: "820 tokens",
    });
  });

  it("formats thousands with one decimal", () => {
    expect(formatTokenCount(8200)).toEqual({
      label: "8.2K",
      ariaLabel: "8,200 tokens",
    });
  });

  it("formats millions compactly", () => {
    expect(formatTokenCount(1_250_000)).toEqual({
      label: "1.3M",
      ariaLabel: "1,250,000 tokens",
    });
  });
});

describe("formatShareOfTaskPct", () => {
  it("formats fractional percentages with one decimal", () => {
    expect(formatShareOfTaskPct(15.4)).toBe("15.4%");
  });

  it("rounds whole percentages", () => {
    expect(formatShareOfTaskPct(100)).toBe("100%");
  });
});
