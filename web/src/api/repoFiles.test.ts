import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { listRepoFiles } from "./repoFiles";

function jsonResponse(body: unknown, status = 200) {
  return new Response(JSON.stringify(body), {
    status,
    headers: { "Content-Type": "application/json" },
  });
}

describe("listRepoFiles", () => {
  beforeEach(() => {
    vi.stubGlobal("fetch", vi.fn());
  });

  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it("sends the worktree and returns the listing", async () => {
    vi.mocked(fetch).mockResolvedValue(
      jsonResponse({ paths: ["a.ts", "b.ts"], truncated: false, source: "git" }),
    );

    await expect(listRepoFiles(" wt-1 ")).resolves.toEqual({
      paths: ["a.ts", "b.ts"],
      truncated: false,
      source: "git",
    });
    expect(String(vi.mocked(fetch).mock.calls[0]?.[0])).toContain(
      "worktree_id=wt-1",
    );
  });

  it("returns null when the repo is not configured", async () => {
    vi.mocked(fetch).mockResolvedValue(new Response("", { status: 409 }));
    await expect(listRepoFiles("wt-1")).resolves.toBeNull();

    vi.mocked(fetch).mockResolvedValue(new Response("", { status: 503 }));
    await expect(listRepoFiles("wt-1")).resolves.toBeNull();
  });

  it("preserves truncation and the walk source", async () => {
    vi.mocked(fetch).mockResolvedValue(
      jsonResponse({ paths: ["a.ts"], truncated: true, source: "walk" }),
    );

    await expect(listRepoFiles("wt-1")).resolves.toMatchObject({
      truncated: true,
      source: "walk",
    });
  });

  it("rejects a blank worktree without a request", async () => {
    await expect(listRepoFiles("   ")).rejects.toThrow(/worktree_id/);
    expect(fetch).not.toHaveBeenCalled();
  });

  it("throws on a malformed payload", async () => {
    vi.mocked(fetch).mockResolvedValue(jsonResponse({ nope: true }));
    await expect(listRepoFiles("wt-1")).rejects.toThrow(/unexpected/);
  });
});
