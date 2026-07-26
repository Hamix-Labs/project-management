import { filterCursorModelsForSelect } from "@/api/cursorModels";
import { useAppSettings } from "@/settings/useAppSettings";
import { useCursorModels } from "@/settings/hooks/useCursorModels";

export function useTaskCreateAgentOptions(runner: string, enabled = true) {
  const { settings } = useAppSettings();
  const cursorBinKey = (settings?.cursor_bin ?? "").trim();
  const modelsEnabled = enabled && runner === "cursor";

  const { query: modelsQuery, modelIds, data } = useCursorModels(
    runner,
    cursorBinKey,
    modelsEnabled,
  );

  const modelsForSelect = filterCursorModelsForSelect(
    data?.ok ? data.models : undefined,
  );

  const modelFetchError = modelsQuery.isError
    ? modelsQuery.error instanceof Error
      ? modelsQuery.error.message
      : String(modelsQuery.error)
    : null;

  const modelServerError =
    data && !data.ok ? (data.error ?? "Model list failed.") : null;

  const workspaceVerifyChatMode =
    settings?.verify_chat_mode === "different_chat"
      ? ("different_chat" as const)
      : ("same_chat" as const);

  return {
    modelIds,
    modelsForSelect,
    modelSelectBusy: modelsQuery.isFetching,
    modelFetchError,
    modelServerError,
    /** Workspace verify-chat default; empty task value inherits this. */
    workspaceVerifyChatMode,
  };
}
