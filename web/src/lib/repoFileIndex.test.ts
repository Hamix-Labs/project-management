import { afterEach, describe, expect, it, vi } from "vitest";
import {
  clearRepoFileIndex,
  getRepoFileIndexSnapshot,
  seedRepoFileIndexForTest,
  warmRepoFileIndex,
} from "./repoFileIndex";

vi.mock("@/api/repoFiles", () => ({
  fetchRepoFilesPage: vi.fn(),
}));

import { fetchRepoFilesPage } from "@/api/repoFiles";

const fetchMock = vi.mocked(fetchRepoFilesPage);

describe("repoFileIndex", () => {
  afterEach(() => {
    clearRepoFileIndex();
    fetchMock.mockReset();
  });

  it("warms pages into a ready index", async () => {
    fetchMock
      .mockResolvedValueOnce({
        paths: ["a.go", "b.go"],
        has_more: true,
        next_after: "b.go",
        source: "git",
      })
      .mockResolvedValueOnce({
        paths: ["c.go"],
        has_more: false,
        source: "git",
      });

    warmRepoFileIndex("wt-1");
    await vi.waitFor(() => {
      expect(getRepoFileIndexSnapshot("wt-1").status).toBe("ready");
    });
    expect(getRepoFileIndexSnapshot("wt-1").paths).toEqual([
      "a.go",
      "b.go",
      "c.go",
    ]);
    expect(fetchMock).toHaveBeenCalledTimes(2);
  });

  it("seed helper sets ready paths for tests", () => {
    seedRepoFileIndexForTest("wt-2", ["x.ts"]);
    expect(getRepoFileIndexSnapshot("wt-2")).toMatchObject({
      status: "ready",
      paths: ["x.ts"],
    });
  });
});
