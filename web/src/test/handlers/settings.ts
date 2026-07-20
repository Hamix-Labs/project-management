import { http, HttpResponse } from "msw";
import type {
  AppSettings,
  AppSettingsPatch,
  ListCursorModelsResult,
  ProbeCursorResult,
} from "@/api/settings";
import { createDeferred } from "@/test/deferred";
import { APP_SETTINGS_DEFAULTS } from "@/test/settingsDefaults";
import { TASK_TEST_DEFAULTS } from "@/test/taskDefaults";

export function appSettingsOk(overrides: Partial<AppSettings> = {}) {
  return http.get("/settings", () =>
    HttpResponse.json({ ...APP_SETTINGS_DEFAULTS, ...overrides }),
  );
}

export function listCursorModelsOk(
  overrides: Partial<ListCursorModelsResult> = {},
) {
  return http.post("/settings/list-cursor-models", () =>
    HttpResponse.json({
      ok: true,
      runner: TASK_TEST_DEFAULTS.runner,
      models: [{ id: "test", label: "Test" }],
      ...overrides,
    }),
  );
}

/** PATCH /settings — success; optional body capture via onPatch. */
export function appSettingsPatchOk(
  response: Partial<AppSettings> = {},
  onPatch?: (body: AppSettingsPatch) => void,
) {
  return http.patch("/settings", async ({ request }) => {
    const body = (await request.json()) as AppSettingsPatch;
    onPatch?.(body);
    return HttpResponse.json({ ...APP_SETTINGS_DEFAULTS, ...response });
  });
}

/** PATCH /settings — 4xx/5xx failure. */
export function appSettingsPatchError(
  status = 500,
  error = "internal: disk full",
) {
  return http.patch("/settings", () =>
    HttpResponse.json({ error }, { status }),
  );
}

/** Keeps PATCH /settings pending until deferred.resolve/reject. */
export function appSettingsPatchPending() {
  const deferred = createDeferred<Response>();
  const handler = http.patch("/settings", () => deferred.promise);
  return [handler, deferred] as const;
}

export function probeCursorOk(result: Partial<ProbeCursorResult> = {}) {
  return http.post("/settings/probe-cursor", () =>
    HttpResponse.json({
      ok: true,
      runner: TASK_TEST_DEFAULTS.runner,
      version: "2026.04",
      ...result,
    }),
  );
}

export function probeCursorFail(error = "spawn ENOENT") {
  return http.post("/settings/probe-cursor", () =>
    HttpResponse.json({
      ok: false,
      runner: TASK_TEST_DEFAULTS.runner,
      error,
    }),
  );
}
