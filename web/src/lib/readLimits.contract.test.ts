import { describe, expect, it } from "vitest";
import contract from "../../../testdata/readlimits.json";
import { READ_LIMITS } from "./readLimits";

describe("readLimits contract parity", () => {
  it("matches testdata/readlimits.json (Go readpolicy contract)", () => {
    for (const key of Object.keys(READ_LIMITS) as (keyof typeof READ_LIMITS)[]) {
      expect(READ_LIMITS[key]).toBe(contract[key]);
    }
  });
});
