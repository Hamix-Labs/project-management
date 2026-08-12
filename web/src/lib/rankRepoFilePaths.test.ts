import { describe, expect, it } from "vitest";
import { rankRepoFilePaths } from "./rankRepoFilePaths";

describe("rankRepoFilePaths", () => {
  const paths = [
    "pkgs/repo/root.go",
    "web/src/App.tsx",
    "README.md",
    "web/src/components/AppShell.tsx",
  ];

  it("returns index order for an empty query", () => {
    expect(rankRepoFilePaths(paths, "")).toEqual(paths);
  });

  it("prefers basename matches", () => {
    const ranked = rankRepoFilePaths(paths, "App");
    expect(ranked[0]).toBe("web/src/App.tsx");
    expect(ranked).toContain("web/src/components/AppShell.tsx");
  });

  it("matches path substrings", () => {
    const ranked = rankRepoFilePaths(paths, "repo");
    expect(ranked).toEqual(["pkgs/repo/root.go"]);
  });
});
