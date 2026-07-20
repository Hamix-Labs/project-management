import { useMemo } from "react";
import { useQuery } from "@tanstack/react-query";
import {
  listCursorModels,
  type ListCursorModelsResult,
} from "@/api/settings";
import { settingsQueryKeys } from "@/lib/settingsQueryKeys";

export function modelIdsFromList(
  data: ListCursorModelsResult | undefined,
): Set<string> {
  if (!data?.ok || !data.models) return new Set<string>();
  return new Set(data.models.map((x) => x.id));
}

type UseCursorModelsOptions = {
  enabled?: boolean;
  queryKey?: readonly unknown[];
};

export function useCursorModels(
  runner: string,
  binaryPath: string,
  options?: boolean | UseCursorModelsOptions,
) {
  const opts: UseCursorModelsOptions =
    typeof options === "boolean" ? { enabled: options } : (options ?? {});
  const trimmedBin = binaryPath.trim();
  const queryKey =
    opts.queryKey ?? settingsQueryKeys.cursorModels(runner, trimmedBin);

  const query = useQuery({
    queryKey,
    queryFn: ({ signal }) =>
      listCursorModels(
        {
          runner,
          binary_path: trimmedBin || undefined,
        },
        { signal },
      ),
    enabled: opts.enabled ?? true,
  });

  const modelIds = useMemo(
    () => modelIdsFromList(query.data),
    [query.data],
  );

  return { query, modelIds, data: query.data };
}
