import { useMutation, useQueryClient } from "@tanstack/react-query";
import { patchProject } from "@/api";
import type { Project, ProjectStatus } from "@/types";
import { invalidateProjectCache } from "./invalidateProjectCache";

type PatchProjectInput = {
  name?: string;
  description?: string;
  status?: ProjectStatus;
  repository_id?: string;
};

export function usePatchProjectMutation(project: Project) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (input: PatchProjectInput) => patchProject(project.id, input),
    onSuccess: (_data, input) => {
      if (
        input.repository_id !== undefined &&
        input.repository_id !== project.repository_id
      ) {
        invalidateProjectCache(queryClient, {
          scope: "repositoryLink",
          projectId: project.id,
          repositoryId: input.repository_id,
        });
        return;
      }
      invalidateProjectCache(queryClient, {
        scope: "detail",
        projectId: project.id,
      });
    },
  });
}
