import { describe, expect, it } from "vitest";
import { buildPatchMutationInput } from "./buildPatchMutationInput";

const base = {
  id: "t1",
  title: "  Hello  ",
  initial_prompt: "<p>body</p>",
  status: "ready" as const,
  priority: "medium" as const,
  project_id: "proj-1" as string | null,
  tagsCsv: "alpha, beta;gamma\ndelta",
  milestone: " m1 ",
  cursor_model: " gpt-4 ",
  pickup_not_before: "2030-01-01T12:00:00Z" as string | null,
};

describe("buildPatchMutationInput", () => {
  it("trims scalars, parses tags like create, and includes schedule when editable", () => {
    expect(buildPatchMutationInput(base)).toEqual({
      id: "t1",
      title: "Hello",
      initial_prompt: "<p>body</p>",
      status: "ready",
      priority: "medium",
      project_id: "proj-1",
      tags: ["alpha", "beta", "gamma", "delta"],
      milestone: "m1",
      cursor_model: "gpt-4",
      pickup_not_before: "2030-01-01T12:00:00Z",
    });
  });

  it("omits pickup_not_before when schedule is not editable for status", () => {
    const input = buildPatchMutationInput({
      ...base,
      status: "running",
    });
    expect(input.pickup_not_before).toBeUndefined();
    expect("pickup_not_before" in input).toBe(false);
  });

  it("nulls empty milestone", () => {
    const input = buildPatchMutationInput({
      ...base,
      project_id: null,
      milestone: "   ",
    });
    expect(input.project_id).toBeNull();
    expect(input.milestone).toBeNull();
  });
});
