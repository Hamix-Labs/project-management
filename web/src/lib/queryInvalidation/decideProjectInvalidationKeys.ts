import { gitQueryKeys } from "@/lib/gitQueryKeys";
import { projectQueryKeys } from "@/lib/projectQueryKeys";
import { taskQueryKeys } from "@/lib/taskQueryKeys";
import type { ProjectInvalidationScope, QueryInvalidationKey } from "./types";

export function decideProjectInvalidationKeys(
  input: ProjectInvalidationScope,
): readonly QueryInvalidationKey[] {
  switch (input.scope) {
    case "list":
      return [projectQueryKeys.all, taskQueryKeys.listRoot()];
    case "detail":
      return [projectQueryKeys.all, projectQueryKeys.detail(input.projectId)];
    case "repositoryLink":
      return [
        projectQueryKeys.all,
        projectQueryKeys.detail(input.projectId),
        gitQueryKeys.projectsByRepo(input.repositoryId),
      ];
  }
}
