/** React Query key tuple returned by catalog functions. */
export type QueryInvalidationKey = readonly unknown[];

export type ProjectInvalidationScope =
  | { scope: "list" }
  | { scope: "detail"; projectId: string }
  | { scope: "context"; projectId: string }
  | {
      scope: "repositoryLink";
      projectId: string;
      repositoryId: string;
    };

export type GitInvalidationScope =
  | { scope: "repositories" }
  | { scope: "repository"; repositoryId: string }
  | { scope: "legacyRepositories"; projectId: string }
  | {
      scope: "legacyRepository";
      projectId: string;
      repositoryId: string;
    };
