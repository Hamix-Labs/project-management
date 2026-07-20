import { describe, expect, it } from "vitest";
import type { GitWorktree } from "@/types/git";
import {
  isFullyRegisteredWorktree,
  isLinkedWorktreeForDisplay,
  pickDefaultWorktreeId,
} from "./gitWorktreeRegistration";

const linked: GitWorktree = {
  id: "00000000-0000-4000-8000-000000000020",
  repository_id: "00000000-0000-4000-8000-000000000010",
  path: "/repo/feature",
  name: "feature",
  is_main: false,
  branch_id: "00000000-0000-4000-8000-000000000030",
  created_at: "2026-06-22T12:00:00Z",
};

const mainSeeded: GitWorktree = {
  ...linked,
  id: "00000000-0000-4000-8000-000000000021",
  path: "/repo/main",
  name: "main",
  is_main: true,
};

const incompleteMainStub: GitWorktree = {
  ...linked,
  id: "00000000-0000-4000-8000-000000000022",
  path: "/repo/main",
  name: "discovered-main",
  is_main: true,
  branch_id: undefined,
};

describe("gitWorktreeRegistration", () => {
  it("treats branch-bound rows as fully registered", () => {
    expect(isFullyRegisteredWorktree(linked)).toBe(true);
    expect(isFullyRegisteredWorktree(incompleteMainStub)).toBe(false);
  });

  it("shows only registered linked worktrees", () => {
    expect(isLinkedWorktreeForDisplay(linked)).toBe(true);
    expect(isLinkedWorktreeForDisplay(mainSeeded)).toBe(false);
    expect(isLinkedWorktreeForDisplay(incompleteMainStub)).toBe(false);
  });

  it("picks a managed worktree, never the primary checkout", () => {
    expect(pickDefaultWorktreeId([linked, mainSeeded])).toBe(linked.id);
    expect(pickDefaultWorktreeId([linked])).toBe(linked.id);
    expect(pickDefaultWorktreeId([mainSeeded, incompleteMainStub])).toBe("");
  });
});
