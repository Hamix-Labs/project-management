export const settingsQueryKeys = {
  all: ["settings"] as const,
  app: () => [...settingsQueryKeys.all, "app"] as const,
  /** Prefix for cursor/verify model list queries (partial-match SSE invalidation). */
  modelsRoot: () => [...settingsQueryKeys.all, "models"] as const,
  /** Settings editor: saved bin, form bin, and runner (legacy key shape). */
  cursorModelsSettings: (
    savedBin: string | undefined,
    formBin: string | undefined,
    runner: string,
  ) =>
    [...settingsQueryKeys.all, "cursor-models", savedBin, formBin, runner] as const,
  /** Shared runner + effective binary path (create modal, etc.). */
  cursorModels: (runner: string, binaryPath: string) =>
    [...settingsQueryKeys.modelsRoot(), "cursor", runner, binaryPath] as const,
};
