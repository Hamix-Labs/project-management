import {
  type Dispatch,
  type FormEvent,
  type SetStateAction,
  useEffect,
  useMemo,
  useState,
} from "react";
import { type UseMutationResult, type UseQueryResult } from "@tanstack/react-query";
import { useLocation } from "react-router-dom";
import {
  type AppSettings,
  type AppSettingsPatch,
  type ListCursorModelsResult,
  type ProbeCursorResult,
} from "@/api/settings";
import { getTimezoneSelectOptions } from "@/shared/time/appTimezone";
import { useAppSettings } from "./useAppSettings";
import { settingsQueryKeys } from "./settingsQueryKeys";
import { useCursorModels } from "./hooks/useCursorModels";
import { SECTION_IDS } from "./sections";
import {
  SETTINGS_SUCCESS_DISMISS_MS,
  diffPatch,
  toFormState,
  type SettingsFormState,
  type SettingsStatus,
} from "./settingsForm";
import {
  buildTimezoneSelectValueSet,
  computeTimezoneDisplayContext,
  parseSettingsNumericValidation,
  resolveVerifyEffectiveRunner,
  type SettingsNumericValidation,
} from "./settingsFormValidation";

const SETTINGS_HASH_TARGETS: ReadonlySet<string> = new Set([
  ...Object.values(SECTION_IDS),
]);

function mergeFormAfterSettingsPatch(
  cur: SettingsFormState | null,
  formAtSubmit: SettingsFormState,
  next: AppSettings,
): SettingsFormState {
  if (cur === null) return toFormState(next);
  const merged: SettingsFormState = { ...cur };
  if (cur.runner === formAtSubmit.runner) {
    merged.runner = next.runner;
  }
  if (cur.cursorBin === formAtSubmit.cursorBin) {
    merged.cursorBin = next.cursor_bin;
  }
  if (cur.cursorModel === formAtSubmit.cursorModel) {
    merged.cursorModel = next.cursor_model;
  }
  if (cur.maxRunDurationSeconds === formAtSubmit.maxRunDurationSeconds) {
    merged.maxRunDurationSeconds = String(next.max_run_duration_seconds);
  }
  if (cur.streamIdleStuckSeconds === formAtSubmit.streamIdleStuckSeconds) {
    merged.streamIdleStuckSeconds = String(next.stream_idle_stuck_seconds);
  }
  if (cur.agentPickupDelaySeconds === formAtSubmit.agentPickupDelaySeconds) {
    merged.agentPickupDelaySeconds = String(next.agent_pickup_delay_seconds);
  }
  if (cur.displayTimezone === formAtSubmit.displayTimezone) {
    merged.displayTimezone = next.display_timezone;
  }
  return merged;
}

function buildCursorProbeSuccessMessage(result: {
  binary_path?: string;
  version?: string;
}): string {
  const bits: string[] = ["Cursor binary OK"];
  if (result.binary_path) bits.push(`at ${result.binary_path}`);
  if (result.version) bits.push(`(version ${result.version})`);
  return `${bits.join(" ")}.`;
}

function resolveProbeDefaultBin(
  form: SettingsFormState,
  result: { binary_path?: string },
): string | null {
  if (result.binary_path && form.cursorBin.trim() === "") {
    return result.binary_path;
  }
  return null;
}

function useAutoDismissSettingsSuccess(
  status: SettingsStatus,
  setStatus: (status: SettingsStatus) => void,
) {
  useEffect(() => {
    if (status?.kind !== "success") return;
    const id = window.setTimeout(() => {
      setStatus(null);
    }, SETTINGS_SUCCESS_DISMISS_MS);
    return () => window.clearTimeout(id);
  }, [status, setStatus]);
}

function useSettingsSectionHashScroll(
  locationHash: string,
  isLoading: boolean,
  form: SettingsFormState | null,
  settings: AppSettings | undefined,
) {
  useEffect(() => {
    if (isLoading || !form || !settings) return;
    const hash = locationHash.replace(/^#/, "");
    if (!hash) return;
    if (!SETTINGS_HASH_TARGETS.has(hash)) return;
    const el = document.getElementById(hash);
    if (!el) return;
    const prefersReduced =
      typeof window.matchMedia === "function" &&
      window.matchMedia("(prefers-reduced-motion: reduce)").matches;
    const run = () => {
      el.scrollIntoView({
        behavior: prefersReduced ? "auto" : "smooth",
        block: "start",
      });
    };
    requestAnimationFrame(() => {
      requestAnimationFrame(run);
    });
  }, [isLoading, form, settings, locationHash]);
}

function useSettingsFormHydration(
  settings: AppSettings | undefined,
  form: SettingsFormState | null,
  setForm: (value: SettingsFormState | null) => void,
) {
  useEffect(() => {
    if (settings && form === null) {
      setForm(toFormState(settings));
    }
  }, [settings, form, setForm]);
}

function useSettingsCursorModelQueries(
  settings: AppSettings | undefined,
  form: SettingsFormState | null,
) {
  const runner = form?.runner ?? settings?.runner ?? "cursor";
  const formBin = form?.cursorBin ?? "";

  const { query: cursorModelsQuery, modelIds: modelIdsFromList } = useCursorModels(
    runner,
    formBin,
    {
      enabled: Boolean(settings && form),
      queryKey: settingsQueryKeys.cursorModelsSettings(
        settings?.cursor_bin,
        form?.cursorBin,
        runner,
      ),
    },
  );

  const verifyEffectiveRunner =
    form && settings ? resolveVerifyEffectiveRunner(form, settings) : "cursor";

  const { query: verifyModelsQuery, modelIds: verifyModelIdsFromList } = useCursorModels(
    verifyEffectiveRunner,
    formBin,
    {
      enabled: Boolean(settings && form),
      queryKey: settingsQueryKeys.verifyModels(verifyEffectiveRunner, form?.cursorBin),
    },
  );

  return {
    cursorModelsQuery,
    modelIdsFromList,
    verifyModelsQuery,
    verifyModelIdsFromList,
  };
}

function settingsErrorMessage(err: unknown): string {
  return err instanceof Error ? err.message : String(err);
}

function createSettingsFieldHandler(
  setForm: Dispatch<SetStateAction<SettingsFormState | null>>,
  resolvedDefaultBin: string | null,
  setResolvedDefaultBin: Dispatch<SetStateAction<string | null>>,
) {
  return function handleField<K extends keyof SettingsFormState>(
    key: K,
    value: SettingsFormState[K],
  ) {
    setForm((cur) => (cur === null ? cur : { ...cur, [key]: value }));
    if (key === "cursorBin" && resolvedDefaultBin !== null) {
      setResolvedDefaultBin(null);
    }
  };
}

async function submitSettingsForm(params: {
  settings: AppSettings;
  form: SettingsFormState;
  numericValidation: SettingsNumericValidation;
  patch: UseMutationResult<AppSettings, Error, AppSettingsPatch, unknown>;
  setForm: Dispatch<SetStateAction<SettingsFormState | null>>;
  setStatus: Dispatch<SetStateAction<SettingsStatus>>;
}): Promise<void> {
  const { settings, form, numericValidation, patch, setForm, setStatus } = params;
  const { maxInvalid, streamIdleInvalid, pickupInvalid } = numericValidation;
  if (maxInvalid || streamIdleInvalid || pickupInvalid) return;
  const body = diffPatch(settings, form);
  if (Object.keys(body).length === 0) return;
  const formAtSubmit = form;
  setStatus(null);
  try {
    const next = await patch.mutateAsync(body);
    setForm((cur) => mergeFormAfterSettingsPatch(cur, formAtSubmit, next));
    setStatus({ kind: "success", message: "Settings saved." });
  } catch (err) {
    setStatus({ kind: "error", message: settingsErrorMessage(err) });
  }
}

async function probeCursorBinary(params: {
  form: SettingsFormState;
  probe: UseMutationResult<
    ProbeCursorResult,
    Error,
    { runner?: string; binary_path?: string },
    unknown
  >;
  setStatus: Dispatch<SetStateAction<SettingsStatus>>;
  setResolvedDefaultBin: Dispatch<SetStateAction<string | null>>;
}): Promise<void> {
  const { form, probe, setStatus, setResolvedDefaultBin } = params;
  setStatus(null);
  try {
    const result = await probe.mutateAsync({
      runner: form.runner.trim() || undefined,
      binary_path: form.cursorBin.trim() || undefined,
    });
    if (result.ok) {
      setStatus({
        kind: "success",
        message: buildCursorProbeSuccessMessage(result),
      });
      setResolvedDefaultBin(resolveProbeDefaultBin(form, result));
      return;
    }
    setStatus({
      kind: "error",
      message: `Cursor binary check failed: ${result.error ?? "unknown error"}`,
    });
  } catch (err) {
    setStatus({ kind: "error", message: settingsErrorMessage(err) });
  }
}

export type SettingsPageLoadedViewProps = {
  form: SettingsFormState;
  status: SettingsStatus;
  resolvedDefaultBin: string | null;
  isDirty: boolean;
  numericValidation: SettingsNumericValidation;
  tzSelectOptions: ReturnType<typeof getTimezoneSelectOptions>;
  browserTz: string;
  showCustomTz: boolean;
  lastUpdated: string;
  lastUpdatedFormatted: string;
  cursorModelsQuery: UseQueryResult<ListCursorModelsResult, Error>;
  modelIdsFromList: Set<string>;
  verifyModelsQuery: UseQueryResult<ListCursorModelsResult, Error>;
  verifyModelIdsFromList: Set<string>;
  patchPending: boolean;
  probePending: boolean;
  onField: ReturnType<typeof createSettingsFieldHandler>;
  onSubmit: (e: FormEvent) => void;
  onProbe: () => void;
  onDiscard: () => void;
};

export function useSettingsPageModel() {
  const location = useLocation();
  const { settings, isLoading, error, patch, probe, refetch } = useAppSettings();
  const [form, setForm] = useState<SettingsFormState | null>(null);
  const [status, setStatus] = useState<SettingsStatus>(null);
  const [resolvedDefaultBin, setResolvedDefaultBin] = useState<string | null>(null);

  useAutoDismissSettingsSuccess(status, setStatus);
  useSettingsSectionHashScroll(location.hash, isLoading, form, settings);
  useSettingsFormHydration(settings, form, setForm);

  const isDirty = useMemo(() => {
    if (!settings || !form) return false;
    return Object.keys(diffPatch(settings, form)).length > 0;
  }, [settings, form]);

  const {
    cursorModelsQuery,
    modelIdsFromList,
    verifyModelsQuery,
    verifyModelIdsFromList,
  } = useSettingsCursorModelQueries(settings, form);

  const tzSelectOptions = useMemo(() => getTimezoneSelectOptions(), []);
  const tzValueSet = useMemo(() => buildTimezoneSelectValueSet(), []);

  const numericValidation = parseSettingsNumericValidation(form);
  const handleField = createSettingsFieldHandler(
    setForm,
    resolvedDefaultBin,
    setResolvedDefaultBin,
  );

  const loadedViewProps = useMemo((): SettingsPageLoadedViewProps | null => {
    if (isLoading || !form || !settings) return null;
    const lastUpdated = settings.updated_at ?? "";
    const { browserTz, lastUpdatedFormatted, showCustomTz } =
      computeTimezoneDisplayContext(form, lastUpdated, tzValueSet);
    return {
      form,
      status,
      resolvedDefaultBin,
      isDirty,
      numericValidation,
      tzSelectOptions,
      browserTz,
      showCustomTz,
      lastUpdated,
      lastUpdatedFormatted,
      cursorModelsQuery,
      modelIdsFromList,
      verifyModelsQuery,
      verifyModelIdsFromList,
      patchPending: patch.isPending,
      probePending: probe.isPending,
      onField: handleField,
      onSubmit: (e) => {
        e.preventDefault();
        void submitSettingsForm({
          settings,
          form,
          numericValidation,
          patch,
          setForm,
          setStatus,
        });
      },
      onProbe: () => {
        void probeCursorBinary({
          form,
          probe,
          setStatus,
          setResolvedDefaultBin,
        });
      },
      onDiscard: () => setForm(toFormState(settings)),
    };
  }, [
    isLoading,
    form,
    settings,
    status,
    resolvedDefaultBin,
    isDirty,
    numericValidation,
    tzSelectOptions,
    tzValueSet,
    cursorModelsQuery,
    modelIdsFromList,
    verifyModelsQuery,
    verifyModelIdsFromList,
    patch,
    probe,
    handleField,
  ]);

  return {
    error,
    refetch,
    loadedViewProps,
  };
}
