import { gitQueryKeys } from "@/lib/gitQueryKeys";
import type { GitInvalidationScope, QueryInvalidationKey } from "./types";

export function decideGitInvalidationKeys(
  input: GitInvalidationScope,
): readonly QueryInvalidationKey[] {
  switch (input.scope) {
    case "repositories":
      return [gitQueryKeys.globalRepositories()];
    case "repository":
      return [
        gitQueryKeys.globalRepositories(),
        gitQueryKeys.globalRepository(input.repositoryId),
        gitQueryKeys.globalWorktrees(input.repositoryId),
        gitQueryKeys.globalBranches(input.repositoryId),
        gitQueryKeys.projectsByRepo(input.repositoryId),
      ];
  }
}
