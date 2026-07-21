import { describe, expect, it } from "vitest";
import { ApiError } from "@/api";
import { gitRegisterErrorMessage, isDuplicateRegisterError } from "./gitRegisterErrors";

describe("gitRegisterErrors", () => {
  it("maps duplicate conflicts to clean copy without request ids", () => {
    const err = new ApiError("repository already registered (request abc)", {
      status: 409,
      code: "duplicate",
      requestId: "abc",
    });
    expect(isDuplicateRegisterError(err)).toBe(true);
    expect(gitRegisterErrorMessage(err)).toBe("This repository is already registered.");
  });

  it("falls back to the error message for other failures", () => {
    const err = new ApiError("not a git repository (request x)", {
      status: 409,
      code: "not_a_git_repository",
      requestId: "x",
    });
    expect(isDuplicateRegisterError(err)).toBe(false);
    expect(gitRegisterErrorMessage(err)).toBe("not a git repository (request x)");
    expect(gitRegisterErrorMessage(new Error("boom"))).toBe("boom");
    expect(gitRegisterErrorMessage("nope")).toBe("Register failed");
  });
});
