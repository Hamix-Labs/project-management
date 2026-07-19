import { describe, expect, it } from "vitest";
import { gitQueryKeys } from "@/lib/gitQueryKeys";
import { decideGitInvalidationKeys } from "./decideGitInvalidationKeys";
import type { GitInvalidationScope } from "./types";

describe("decideGitInvalidationKeys", () => {
  const cases: {
    name: string;
    input: GitInvalidationScope;
    expected: readonly (readonly unknown[])[];
  }[] = [
    {
      name: "repositories",
      input: { scope: "repositories" },
      expected: [gitQueryKeys.globalRepositories()],
    },
    {
      name: "repository",
      input: { scope: "repository", repositoryId: "repo-1" },
      expected: [
        gitQueryKeys.globalRepositories(),
        gitQueryKeys.globalRepository("repo-1"),
        gitQueryKeys.globalWorktrees("repo-1"),
        gitQueryKeys.globalBranches("repo-1"),
        gitQueryKeys.globalWorktreeCheckoutStatus("repo-1"),
        gitQueryKeys.projectsByRepo("repo-1"),
      ],
    },
  ];

  it.each(cases)("$name returns the catalog keys", ({ input, expected }) => {
    expect(decideGitInvalidationKeys(input)).toEqual(expected);
  });
});
