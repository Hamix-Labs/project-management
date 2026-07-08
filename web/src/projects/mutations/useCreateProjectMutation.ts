import { useMutation, useQueryClient } from "@tanstack/react-query";
import { createProject } from "@/api";
import { invalidateProjectCache } from "./invalidateProjectCache";

export function useCreateProjectMutation() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: createProject,
    onSuccess: (project, variables) => {
      if (variables.repository_id) {
        invalidateProjectCache(
          queryClient,
          { scope: "list" },
          {
            scope: "repositoryLink",
            projectId: project.id,
            repositoryId: variables.repository_id,
          },
        );
        return;
      }
      invalidateProjectCache(queryClient, { scope: "list" });
    },
  });
}
