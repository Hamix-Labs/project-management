import { describe, expect, it } from "vitest";
import { repositoryDisplayName } from "./repositoryDisplayName";

describe("repositoryDisplayName", () => {
  it("returns the last path segment", () => {
    expect(repositoryDisplayName("C:/Users/dev/Documents/hamix")).toBe("hamix");
    expect(repositoryDisplayName("/repo/main")).toBe("main");
  });

  it("handles trailing separators and backslashes", () => {
    expect(repositoryDisplayName("C:\\Users\\dev\\Documents\\hamix\\")).toBe("hamix");
  });

  it("returns the original string when empty after trim", () => {
    expect(repositoryDisplayName("   ")).toBe("   ");
  });
});
