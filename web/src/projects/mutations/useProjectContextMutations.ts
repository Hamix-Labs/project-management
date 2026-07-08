import { useMutation, useQueryClient } from "@tanstack/react-query";
import {
  createProjectContext,
  createProjectContextEdge,
  deleteProjectContext,
  deleteProjectContextEdge,
  patchProjectContext,
  patchProjectContextEdge,
} from "@/api";
import type {
  ProjectContextKind,
  ProjectContextRelation,
} from "@/types";
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
  const createEdgeMutation = useMutation({
    mutationFn: (input: {
      source_context_id: string;
      target_context_id: string;
      relation: ProjectContextRelation;
      strength: number;
      note: string;
    }) => createProjectContextEdge(projectId, input),
    onSuccess: invalidate,
  });
  const patchEdgeMutation = useMutation({
    mutationFn: (input: {
      id: string;
      relation: ProjectContextRelation;
      strength: number;
      note: string;
    }) => {
      const { id, ...patch } = input;
      return patchProjectContextEdge(projectId, id, patch);
    },
    onSuccess: invalidate,
  });
  const deleteEdgeMutation = useMutation({
    mutationFn: (edgeId: string) => deleteProjectContextEdge(projectId, edgeId),
    onSuccess: invalidate,
  });

  return {
    createContextMutation,
    patchContextMutation,
    deleteContextMutation,
    createEdgeMutation,
    patchEdgeMutation,
    deleteEdgeMutation,
  };
}
