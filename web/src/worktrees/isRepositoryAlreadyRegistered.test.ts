import { describe, expect, it } from "vitest";
import { isRepositoryAlreadyRegistered } from "./isRepositoryAlreadyRegistered";

describe("isRepositoryAlreadyRegistered", () => {
  const registered = [
    { path: "/repos/hamix", host_path: "C:/Users/dev/Documents/hamix" },
    { path: "/repos/other", host_path: "" },
  ];

  it("returns false for an empty candidate", () => {
    expect(isRepositoryAlreadyRegistered("", registered)).toBe(false);
    expect(isRepositoryAlreadyRegistered("   ", registered)).toBe(false);
  });

  it("returns false when nothing is registered", () => {
    expect(isRepositoryAlreadyRegistered("/repos/hamix", [])).toBe(false);
  });

  it("matches a registered path ignoring separators and case", () => {
    expect(isRepositoryAlreadyRegistered("/repos/hamix", registered)).toBe(true);
    expect(isRepositoryAlreadyRegistered("\\repos\\HAMIX\\", registered)).toBe(true);
  });

  it("matches a registered host_path", () => {
    expect(
      isRepositoryAlreadyRegistered("C:/Users/dev/Documents/hamix", registered),
    ).toBe(true);
    expect(
      isRepositoryAlreadyRegistered("C:\\Users\\dev\\Documents\\hamix", registered),
    ).toBe(true);
  });

  it("returns false for an unregistered path", () => {
    expect(isRepositoryAlreadyRegistered("/repos/new", registered)).toBe(false);
  });
});
