import { describe, expect, it } from "vitest";
import { browseOrderPaths, rankMentionPaths } from "./promptFileMentionRank";

describe("browseOrderPaths", () => {
  it("puts dot-paths last instead of first", () => {
    const ordered = browseOrderPaths([
      ".codegraph/index.db",
      ".cursor/rules/plan.mdc",
      "web/src/main.tsx",
      "README.md",
    ]);

    expect(ordered.at(-1)).toMatch(/^\./);
    expect(ordered.at(-2)).toMatch(/^\./);
    expect(ordered.slice(0, 2)).toEqual(["README.md", "web/src/main.tsx"]);
  });

  it("prefers shallower paths, then alphabetical", () => {
    expect(
      browseOrderPaths(["a/b/c.ts", "z.ts", "a/b.ts", "a.ts"]),
    ).toEqual(["a.ts", "z.ts", "a/b.ts", "a/b/c.ts"]);
  });

  it("returns every path", () => {
    const paths = Array.from({ length: 500 }, (_, i) => `file-${i}.ts`);
    expect(browseOrderPaths(paths)).toHaveLength(500);
  });
});

describe("rankMentionPaths", () => {
  it("falls back to browse order for an empty query", () => {
    expect(rankMentionPaths([".git-hook", "b.ts", "a.ts"], "  ")).toEqual([
      "a.ts",
      "b.ts",
      ".git-hook",
    ]);
  });

  it("ranks a file name match above a directory-only match", () => {
    const ranked = rankMentionPaths(
      ["mention/PromptEditorMenu.tsx", "web/src/mention.ts"],
      "mention",
    );

    expect(ranked[0]).toBe("web/src/mention.ts");
  });

  it("sinks dot-paths below real matches", () => {
    const ranked = rankMentionPaths(
      [".cursor/rules/test.mdc", "web/src/taskTest.ts"],
      "test",
    );

    expect(ranked[0]).toBe("web/src/taskTest.ts");
    expect(ranked.at(-1)).toBe(".cursor/rules/test.mdc");
  });

  it("matches a path fragment typed with separators", () => {
    const ranked = rankMentionPaths(
      ["web/src/main.tsx", "docs/api.md", "pkgs/repo/root.go"],
      "web/src main",
    );

    expect(ranked).toEqual(["web/src/main.tsx"]);
  });

  it("returns every match rather than a capped page", () => {
    const paths = Array.from({ length: 300 }, (_, i) => `src/widget-${i}.ts`);

    expect(rankMentionPaths(paths, "widget")).toHaveLength(300);
  });

  it("returns nothing when the query matches nothing", () => {
    expect(rankMentionPaths(["a.ts", "b.ts"], "zzzznope")).toEqual([]);
  });
});
