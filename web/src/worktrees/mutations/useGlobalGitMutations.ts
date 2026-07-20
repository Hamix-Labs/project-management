import { useMutation, useQueryClient } from "@tanstack/react-query";
import {
  createGlobalGitRepository,
  deleteGlobalGitRepository,
} from "@/api/gitGlobal";
import {
  invalidateGitCache,
} from "./invalidateGitCache";

export function useGlobalGitMutations() {
  const qc = useQueryClient();

  const createRepository = useMutation({
    mutationFn: createGlobalGitRepository,
    onSuccess: () => {
      invalidateGitCache(qc, { scope: "repositories" });
    },
  });

  const deleteRepository = useMutation({
    mutationFn: deleteGlobalGitRepository,
    onSuccess: () => {
      invalidateGitCache(qc, { scope: "repositories" });
    },
  });

  return {
    createRepository,
    deleteRepository,
  };
}
