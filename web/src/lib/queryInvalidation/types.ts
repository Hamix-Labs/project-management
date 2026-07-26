/** React Query key tuple returned by catalog functions. */
export type QueryInvalidationKey = readonly unknown[];

export type ProjectInvalidationScope =
  | { scope: "list" }
  | { scope: "detail"; projectId: string }
  | {
      scope: "repositoryLink";
      projectId: string;
      repositoryId: string;
    };

export type GitInvalidationScope =
  | { scope: "repositories" }
  | { scope: "repository"; repositoryId: string };

export type TaskInvalidationScope =
  | { scope: "listStats" }
  | { scope: "detail"; taskId: string }
  | { scope: "checklist"; taskId: string }
  | { scope: "events"; taskId: string }
  | { scope: "drafts" }
  | { scope: "templates" };
