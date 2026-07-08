import { useQuery } from "@tanstack/react-query";
import { useGlobalRepositories } from "@/hooks/useGlobalRepositories";
import { gitQueryKeys } from "@/lib/gitQueryKeys";
import { useProject } from "@/projects/hooks";
import { resolveTaskGitBinding } from "../task-git/resolveTaskGitBinding";

export function useTaskGitBinding(
  worktreeId: string | undefined,
  projectId?: string,
) {
  const wtId = (worktreeId ?? "").trim();
  const projectKey = (projectId ?? "").trim();

  const repositoriesQuery = useGlobalRepositories({
    enabled: wtId !== "",
  });
  const repositories = repositoriesQuery.data ?? [];

  const projectQuery = useProject(projectKey, {
    enabled: wtId !== "" && projectKey !== "",
  });
  const repositoryIdHint = projectQuery.data?.repository_id?.trim() || undefined;

  const repositoriesReady =
    !repositoriesQuery.isLoading && repositories.length > 0;
  const projectReady = projectKey === "" || !projectQuery.isLoading;

  return useQuery({
    queryKey: gitQueryKeys.taskBinding(wtId, repositoryIdHint),
    queryFn: ({ signal }) =>
      resolveTaskGitBinding(wtId, repositories, {
        repositoryIdHint,
        signal,
      }),
    enabled: wtId !== "" && repositoriesReady && projectReady,
    staleTime: 60_000,
  });
}
