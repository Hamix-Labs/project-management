import {
  useQuery,
  type QueryFunction,
  type QueryKey,
  type UseQueryOptions,
} from "@tanstack/react-query";

type RepoScopedQueryOptions<TData, TQueryKey extends QueryKey> = {
  enabled?: boolean;
} & Omit<
  UseQueryOptions<TData, Error, TData, TQueryKey>,
  "queryKey" | "queryFn" | "enabled"
>;

type RepoScopedQueryArgs<TData, TQueryKey extends QueryKey> = {
  repositoryId: string;
  queryKey: TQueryKey;
  queryFn: QueryFunction<TData, TQueryKey>;
  options?: RepoScopedQueryOptions<TData, TQueryKey>;
};

/** Centralizes the repo-id enable gate shared by git/project repository-scoped queries. */
export function useRepoScopedQuery<TData, TQueryKey extends QueryKey = QueryKey>({
  repositoryId,
  queryKey,
  queryFn,
  options,
}: RepoScopedQueryArgs<TData, TQueryKey>) {
  const { enabled, ...queryOptions } = options ?? {};
  return useQuery({
    queryKey,
    queryFn,
    enabled: enabled !== false && repositoryId.trim() !== "",
    ...queryOptions,
  });
}
