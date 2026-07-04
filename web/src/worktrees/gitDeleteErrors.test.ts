import { describe, expect, it } from "vitest";
import { ApiError } from "@/api";
import { gitDeleteNeedsForce, gitDeleteErrorMessage } from "./gitDeleteErrors";

describe("gitDeleteErrors", () => {
  it("detects dirty worktree errors that need force", () => {
    const err = new ApiError("worktree has uncommitted changes; use force", {
      status: 409,
      code: "path_exists",
    });
    expect(gitDeleteNeedsForce(err)).toBe(true);
    expect(gitDeleteErrorMessage(err)).toMatch(/force remove/i);
  });
});
