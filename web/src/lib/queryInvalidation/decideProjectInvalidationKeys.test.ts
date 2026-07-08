import { describe, expect, it } from "vitest";
import { gitQueryKeys } from "@/lib/gitQueryKeys";
import { projectQueryKeys } from "@/lib/projectQueryKeys";
import { taskQueryKeys } from "@/lib/taskQueryKeys";
import { decideProjectInvalidationKeys } from "./decideProjectInvalidationKeys";
import type { ProjectInvalidationScope } from "./types";

describe("decideProjectInvalidationKeys", () => {
  const cases: {
    name: string;
    input: ProjectInvalidationScope;
    expected: readonly (readonly unknown[])[];
  }[] = [
    {
      name: "list",
      input: { scope: "list" },
      expected: [projectQueryKeys.all, taskQueryKeys.listRoot()],
    },
    {
      name: "detail",
      input: { scope: "detail", projectId: "proj-1" },
      expected: [projectQueryKeys.all, projectQueryKeys.detail("proj-1")],
    },
    {
      name: "context",
      input: { scope: "context", projectId: "proj-1" },
      expected: [
        projectQueryKeys.context("proj-1"),
        projectQueryKeys.detail("proj-1"),
      ],
    },
    {
      name: "repositoryLink",
      input: {
        scope: "repositoryLink",
        projectId: "proj-1",
        repositoryId: "repo-1",
      },
      expected: [
        projectQueryKeys.all,
        projectQueryKeys.detail("proj-1"),
        gitQueryKeys.projectsByRepo("repo-1"),
      ],
    },
  ];

  it.each(cases)("$name returns the catalog keys", ({ input, expected }) => {
    expect(decideProjectInvalidationKeys(input)).toEqual(expected);
  });
});
