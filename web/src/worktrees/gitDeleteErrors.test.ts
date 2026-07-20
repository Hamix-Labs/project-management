import { describe, expect, it } from "vitest";
import { ApiError } from "@/api";
import { gitDeleteBlocked, gitDeleteErrorMessage } from "./gitDeleteErrors";

describe("gitDeleteErrors", () => {
  it("maps running-task conflicts to a blocked delete", () => {
    const err = new ApiError("A task is still running", {
      status: 409,
      code: "has_running_task",
    });
    expect(gitDeleteBlocked(err)).toBe(true);
    expect(gitDeleteErrorMessage(err)).toMatch(/still running/i);
  });

  it("falls back to the error message for other failures", () => {
    expect(gitDeleteErrorMessage(new Error("boom"))).toBe("boom");
    expect(gitDeleteErrorMessage("nope")).toBe("Delete failed");
  });
});
