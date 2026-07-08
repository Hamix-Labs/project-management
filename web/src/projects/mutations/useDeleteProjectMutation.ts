import { useMutation, useQueryClient } from "@tanstack/react-query";
import { deleteProject } from "@/api";
import { invalidateProjectCache } from "./invalidateProjectCache";

export function useDeleteProjectMutation(projectId: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: () => deleteProject(projectId),
    onSuccess: () => {
      invalidateProjectCache(queryClient, { scope: "list" });
    },
  });
}
