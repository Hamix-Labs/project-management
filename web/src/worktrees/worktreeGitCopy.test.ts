import { describe, expect, it } from "vitest";
import {
  unregisterWorktreeAriaLabel,
  worktreeAriaLabel,
  worktreeCountLabel,
} from "./worktreeGitCopy";

describe("worktreeGitCopy helpers", () => {
  it("formats worktree counts", () => {
    expect(worktreeCountLabel(1)).toBe("1 worktree");
    expect(worktreeCountLabel(2)).toBe("2 worktrees");
  });

  it("formats aria labels", () => {
    expect(worktreeAriaLabel("feature")).toBe("Worktree: feature");
    expect(unregisterWorktreeAriaLabel("feature")).toBe('Unregister worktree "feature"');
  });
});
