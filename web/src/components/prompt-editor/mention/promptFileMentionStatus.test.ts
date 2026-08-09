import { describe, expect, it } from "vitest";
import {
  describeMentionSearchStatus,
  mentionStatusFromHttpStatus,
  type MentionSearchStatus,
} from "./promptFileMentionStatus";

describe("describeMentionSearchStatus", () => {
  it("stays silent when there is nothing to report", () => {
    expect(describeMentionSearchStatus({ kind: "idle" })).toBeNull();
    expect(
      describeMentionSearchStatus({
        kind: "ready",
        matched: 12,
        truncated: false,
      }),
    ).toBeNull();
  });

  it("admits when the list it is searching is partial", () => {
    const hint = describeMentionSearchStatus({
      kind: "ready",
      matched: 12,
      truncated: true,
    });

    expect(hint?.tone).toBe("info");
    expect(hint?.message).toContain("50,000");
  });

  it("says an empty repository is empty, not unmatched", () => {
    const hint = describeMentionSearchStatus({ kind: "empty-repo" });

    expect(hint?.tone).toBe("info");
    expect(hint?.message).toMatch(/no files/i);
  });

  it("separates a missing worktree from a broken one", () => {
    const missing = describeMentionSearchStatus({ kind: "worktree-missing" });
    const broken = describeMentionSearchStatus({ kind: "worktree-broken" });

    expect(missing?.message).toMatch(/no longer exists/i);
    expect(broken?.message).toMatch(/missing or not a directory/i);
    expect(missing?.message).not.toBe(broken?.message);
  });

  it("does not blame file search when nothing was searched", () => {
    const unbound = describeMentionSearchStatus({ kind: "unbound" });
    const resolving = describeMentionSearchStatus({ kind: "resolving" });

    expect(unbound?.tone).toBe("info");
    expect(unbound?.message).not.toMatch(/failed/i);
    expect(resolving?.tone).toBe("pending");
    expect(resolving?.message).not.toMatch(/failed/i);
  });

  it("only reports a failure for a request that actually failed", () => {
    const failed = describeMentionSearchStatus({ kind: "failed", status: 502 });
    expect(failed?.tone).toBe("error");
    expect(failed?.message).toMatch(/File search failed/i);
  });

  it("points misconfiguration at the repositories page", () => {
    const withLinks: MentionSearchStatus[] = [
      { kind: "no-repo" },
      { kind: "no-main-worktree" },
      { kind: "worktree-missing" },
      { kind: "worktree-broken" },
      { kind: "failed" },
    ];

    for (const status of withLinks) {
      expect(describeMentionSearchStatus(status)?.action).toEqual({
        label: "Repositories page",
        href: "/repositories",
      });
    }
  });

  it("tells the user a timeout is retryable", () => {
    const hint = describeMentionSearchStatus({ kind: "timed-out" });
    expect(hint?.tone).toBe("error");
    expect(hint?.message).toMatch(/reopen the menu/i);
  });
});

describe("mentionStatusFromHttpStatus", () => {
  it("maps 404 to a deleted worktree and 500 to a broken path", () => {
    expect(mentionStatusFromHttpStatus(404)).toEqual({ kind: "worktree-missing" });
    expect(mentionStatusFromHttpStatus(500)).toEqual({ kind: "worktree-broken" });
  });

  it("keeps other statuses as a generic failure", () => {
    expect(mentionStatusFromHttpStatus(502)).toEqual({
      kind: "failed",
      status: 502,
    });
    expect(mentionStatusFromHttpStatus(undefined)).toEqual({
      kind: "failed",
      status: undefined,
    });
  });
});
