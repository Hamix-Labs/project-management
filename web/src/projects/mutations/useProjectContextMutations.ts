import { useMutation, useQueryClient } from "@tanstack/react-query";
import {
  createProjectContext,
  deleteProjectContext,
  patchProjectContext,
} from "@/api";
import type { ProjectContextKind } from "@/types";
import { invalidateProjectCache } from "./invalidateProjectCache";

export function useProjectContextMutations(projectId: string) {
  const queryClient = useQueryClient();
  const invalidate = () => {
    invalidateProjectCache(queryClient, { scope: "context", projectId });
  };

  const createContextMutation = useMutation({
    mutationFn: (input: {
      kind: ProjectContextKind;
      title: string;
      body: string;
      pinned: boolean;
    }) => createProjectContext(projectId, input),
    onSuccess: invalidate,
  });
  const patchContextMutation = useMutation({
    mutationFn: (input: {
      id: string;
      kind: ProjectContextKind;
      title: string;
      body: string;
      pinned: boolean;
    }) => {
      const { id, ...patch } = input;
      return patchProjectContext(projectId, id, patch);
    },
    onSuccess: invalidate,
  });
  const deleteContextMutation = useMutation({
    mutationFn: (contextId: string) => deleteProjectContext(projectId, contextId),
    onSuccess: invalidate,
  });

  return {
    createContextMutation,
    patchContextMutation,
    deleteContextMutation,
  };
}
