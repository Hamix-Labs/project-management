import { describe, expect, it } from "vitest";
import { isAbortError } from "./isAbortError";

describe("isAbortError", () => {
  it("recognizes DOMException AbortError", () => {
    expect(isAbortError(new DOMException("Aborted", "AbortError"))).toBe(true);
  });

  it("recognizes Error with AbortError name", () => {
    expect(isAbortError(Object.assign(new Error("Aborted"), { name: "AbortError" }))).toBe(
      true,
    );
  });

  it("rejects unrelated errors", () => {
    expect(isAbortError(new Error("network"))).toBe(false);
    expect(isAbortError(null)).toBe(false);
  });
});
