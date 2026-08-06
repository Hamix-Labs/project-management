import { describe, expect, it, vi, beforeEach, afterEach } from "vitest";
import { getGlobalGitRepository, listGlobalGitRepositories } from "./gitGlobal";
import { respondGlobalGitApi } from "@/test/handlers/gitGlobal";
import { FACTORY_GIT_REPO_ID } from "@/test/factories/git";

describe("gitGlobal API", () => {
  beforeEach(() => {
    vi.stubGlobal(
      "fetch",
      vi.fn(async (input: RequestInfo | URL, init?: RequestInit) => {
        const url = typeof input === "string" ? input : input.toString();
        const method = init?.method ?? "GET";
        const mocked = respondGlobalGitApi(url, method);
        if (mocked) return mocked;
        return new Response(JSON.stringify({ error: "not found" }), { status: 404 });
      }),
    );
  });

  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it("lists global repositories", async () => {
    const rows = await listGlobalGitRepositories();
    expect(rows).toHaveLength(1);
    expect(rows[0]?.path).toBe("/repo/main");
  });

  it("gets a global repository by id", async () => {
    const repo = await getGlobalGitRepository(FACTORY_GIT_REPO_ID);
    expect(repo.id).toBe(FACTORY_GIT_REPO_ID);
    expect(repo.path).toBe("/repo/main");
  });
});
