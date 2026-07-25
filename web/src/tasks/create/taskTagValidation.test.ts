import { describe, expect, it } from "vitest";
import { ApiError } from "@/api";
import {
  TASK_TAG_RULES_HINT,
  taskMutationErrorMessage,
  validateTagsCsv,
} from "./taskTagValidation";

describe("validateTagsCsv", () => {
  it("accepts empty and valid tags", () => {
    expect(validateTagsCsv("")).toBeNull();
    expect(validateTagsCsv("backend, api")).toBeNull();
    expect(validateTagsCsv("Backend")).toBeNull();
  });

  it("rejects tags with spaces and explains the rules", () => {
    const msg = validateTagsCsv("a a a a a");
    expect(msg).toMatch(/Tag "a a a a a" is invalid/);
    expect(msg).toContain(TASK_TAG_RULES_HINT);
  });

  it("rejects oversized tags", () => {
    const tooLong = "a".repeat(33);
    expect(validateTagsCsv(tooLong)).toMatch(/is invalid/);
  });
});

describe("taskMutationErrorMessage", () => {
  it("rewrites invalid tag API errors and strips request ids", () => {
    const err = new ApiError(
      'invalid tag "a a a a a" (request f25133d1-f58f-4362-82e4-aad920e79fdf)',
      { status: 400, requestId: "f25133d1-f58f-4362-82e4-aad920e79fdf" },
    );
    const msg = taskMutationErrorMessage(err);
    expect(msg).toMatch(/Tag "a a a a a" is invalid/);
    expect(msg).toContain(TASK_TAG_RULES_HINT);
    expect(msg).not.toMatch(/request /i);
  });

  it("uses server detail when present", () => {
    const err = new ApiError(
      'invalid tag "x y": use lowercase letters, numbers, and . _ - only (max 32 characters, no spaces)',
      { status: 400 },
    );
    expect(taskMutationErrorMessage(err)).toBe(
      'Tag "x y" is invalid. use lowercase letters, numbers, and . _ - only (max 32 characters, no spaces)',
    );
  });

  it("rewrites duplicate tag errors", () => {
    expect(
      taskMutationErrorMessage(new Error('duplicate tag "api" (request abc)')),
    ).toBe('Tag "api" is listed more than once.');
  });
});
