import { fetchWithTimeout, jsonHeaders, apiErrorFromResponse } from "./shared";
import {
  isRecord,
  parseBooleanField,
  parseFiniteNumber,
  parseNonEmptyString,
  parseString,
} from "./parseTaskApiCore";

// ---------------------------------------------------------------------------
// Types
// ---------------------------------------------------------------------------

const RUNNER_CONFIG_FIELD_TYPES = [
  "string",
  "secret",
  "int",
  "bool",
  "enum",
] as const;

export type RunnerConfigField = {
  key: string;
  label: string;
  type: (typeof RUNNER_CONFIG_FIELD_TYPES)[number];
  default?: unknown;
  help?: string;
  required?: boolean;
  sensitive?: boolean;
  enum_values?: Array<{ value: string; label: string }>;
};

export type RunnerConfigSchema = {
  version: number;
  fields: RunnerConfigField[];
};

export type RunnerDescriptor = {
  id: string;
  label: string;
  default_binary_hint: string;
  config_schema?: RunnerConfigSchema;
};

export type RunnerProbeResult = {
  ok: boolean;
  runner: string;
  binary_path?: string;
  version?: string;
  error?: string;
};

export type RunnerListModelsResult = {
  ok: boolean;
  runner: string;
  binary_path?: string;
  models?: Array<{ id: string; label: string }>;
  error?: string;
};

export type RunnerValidateConfigResult = {
  valid: boolean;
  error?: string;
};

// ---------------------------------------------------------------------------
// Fetchers
// ---------------------------------------------------------------------------

export async function fetchRunners(
  options?: { signal?: AbortSignal },
): Promise<RunnerDescriptor[]> {
  const res = await fetchWithTimeout("/runners", {
    headers: { Accept: "application/json" },
    signal: options?.signal,
  });
  if (!res.ok) throw await apiErrorFromResponse(res);
  const raw: unknown = await res.json();
  if (!Array.isArray(raw)) {
    throw new Error("Invalid API response: /runners must be an array");
  }
  return raw.map((item, i) => parseRunnerDescriptor(item, `runners[${i}]`));
}

export async function fetchRunnerConfigSchema(
  runnerId: string,
  options?: { signal?: AbortSignal },
): Promise<RunnerConfigSchema> {
  const res = await fetchWithTimeout(`/runners/${encodeURIComponent(runnerId)}/config-schema`, {
    headers: { Accept: "application/json" },
    signal: options?.signal,
  });
  if (!res.ok) throw await apiErrorFromResponse(res);
  const raw: unknown = await res.json();
  return parseConfigSchema(raw);
}

export async function probeRunner(
  runnerId: string,
  body?: { binary_path?: string },
  options?: { signal?: AbortSignal },
): Promise<RunnerProbeResult> {
  const res = await fetchWithTimeout(`/runners/${encodeURIComponent(runnerId)}/probe`, {
    method: "POST",
    headers: jsonHeaders,
    body: body ? JSON.stringify(body) : undefined,
    signal: options?.signal,
  });
  if (!res.ok && res.status !== 404 && res.status !== 501) {
    throw await apiErrorFromResponse(res);
  }
  const raw: unknown = await res.json();
  // Soft probe: 404/501 and partial bodies map to ok:false rather than throw —
  // operators use this as a health check, not CRUD contract enforcement.
  return parseProbeResult(raw);
}

export async function listRunnerModels(
  runnerId: string,
  body?: { binary_path?: string },
  options?: { signal?: AbortSignal },
): Promise<RunnerListModelsResult> {
  const res = await fetchWithTimeout(`/runners/${encodeURIComponent(runnerId)}/list-models`, {
    method: "POST",
    headers: jsonHeaders,
    body: body ? JSON.stringify(body) : undefined,
    signal: options?.signal,
  });
  if (!res.ok && res.status !== 404 && res.status !== 501) {
    throw await apiErrorFromResponse(res);
  }
  const raw: unknown = await res.json();
  // Soft list-models: same intentional soft probe as probeRunner.
  return parseListModelsResult(raw);
}

export async function validateRunnerConfig(
  runnerId: string,
  config: Record<string, unknown>,
  options?: { signal?: AbortSignal },
): Promise<RunnerValidateConfigResult> {
  const res = await fetchWithTimeout(
    `/runners/${encodeURIComponent(runnerId)}/validate-config`,
    {
      method: "POST",
      headers: jsonHeaders,
      body: JSON.stringify(config),
      signal: options?.signal,
    },
  );
  if (!res.ok && res.status !== 422) {
    throw await apiErrorFromResponse(res);
  }
  const raw: unknown = await res.json();
  if (!isRecord(raw)) {
    throw new Error("Invalid API response: validate-config must be an object");
  }
  const out: RunnerValidateConfigResult = {
    valid: parseBooleanField(raw.valid, "valid"),
  };
  if (raw.error !== undefined && raw.error !== null) {
    out.error = parseString(raw.error, "error");
  }
  return out;
}

// ---------------------------------------------------------------------------
// Parsers
// ---------------------------------------------------------------------------

function parseRunnerConfigFieldType(
  value: unknown,
  field: string,
): RunnerConfigField["type"] {
  if (
    typeof value !== "string" ||
    !(RUNNER_CONFIG_FIELD_TYPES as readonly string[]).includes(value)
  ) {
    throw new Error(
      `Invalid API response: ${field} must be one of ${RUNNER_CONFIG_FIELD_TYPES.join(", ")}`,
    );
  }
  return value as RunnerConfigField["type"];
}

function parseRunnerDescriptor(raw: unknown, path: string): RunnerDescriptor {
  if (!isRecord(raw)) {
    throw new Error(`Invalid API response: ${path} must be an object`);
  }
  const out: RunnerDescriptor = {
    id: parseNonEmptyString(raw.id, `${path}.id`),
    label: parseNonEmptyString(raw.label, `${path}.label`),
    default_binary_hint:
      raw.default_binary_hint === undefined || raw.default_binary_hint === null
        ? ""
        : parseString(raw.default_binary_hint, `${path}.default_binary_hint`),
  };
  if (raw.config_schema !== undefined && raw.config_schema !== null) {
    out.config_schema = parseConfigSchema(raw.config_schema, `${path}.config_schema`);
  }
  return out;
}

function parseConfigSchema(
  raw: unknown,
  path = "config_schema",
): RunnerConfigSchema {
  if (!isRecord(raw)) {
    throw new Error(`Invalid API response: ${path} must be an object`);
  }
  if (!Array.isArray(raw.fields)) {
    throw new Error(`Invalid API response: ${path}.fields must be an array`);
  }
  return {
    version: parseFiniteNumber(raw.version, `${path}.version`),
    fields: raw.fields.map((item, i) =>
      parseConfigField(item, `${path}.fields[${i}]`),
    ),
  };
}

function parseConfigField(raw: unknown, path: string): RunnerConfigField {
  if (!isRecord(raw)) {
    throw new Error(`Invalid API response: ${path} must be an object`);
  }
  const field: RunnerConfigField = {
    key: parseNonEmptyString(raw.key, `${path}.key`),
    label: parseNonEmptyString(raw.label, `${path}.label`),
    type: parseRunnerConfigFieldType(raw.type, `${path}.type`),
  };
  if (raw.default !== undefined) field.default = raw.default;
  if (raw.help !== undefined && raw.help !== null) {
    field.help = parseString(raw.help, `${path}.help`);
  }
  if (raw.required !== undefined && raw.required !== null) {
    field.required = parseBooleanField(raw.required, `${path}.required`);
  }
  if (raw.sensitive !== undefined && raw.sensitive !== null) {
    field.sensitive = parseBooleanField(raw.sensitive, `${path}.sensitive`);
  }
  if (raw.enum_values !== undefined && raw.enum_values !== null) {
    if (!Array.isArray(raw.enum_values)) {
      throw new Error(`Invalid API response: ${path}.enum_values must be an array`);
    }
    field.enum_values = raw.enum_values.map((v, i) => {
      if (!isRecord(v)) {
        throw new Error(
          `Invalid API response: ${path}.enum_values[${i}] must be an object`,
        );
      }
      return {
        value: parseNonEmptyString(v.value, `${path}.enum_values[${i}].value`),
        label: parseNonEmptyString(v.label, `${path}.enum_values[${i}].label`),
      };
    });
  }
  return field;
}

/** Soft health-probe body: require object; coerce missing ok/runner rather than throw. */
function parseProbeResult(raw: unknown): RunnerProbeResult {
  if (!isRecord(raw)) {
    throw new Error("Invalid API response: probe response must be an object");
  }
  const out: RunnerProbeResult = {
    ok: typeof raw.ok === "boolean" ? raw.ok : false,
    runner: typeof raw.runner === "string" ? raw.runner : "",
  };
  if (typeof raw.binary_path === "string") out.binary_path = raw.binary_path;
  if (typeof raw.version === "string") out.version = raw.version;
  if (typeof raw.error === "string") out.error = raw.error;
  return out;
}

/** Soft list-models body: same intentional soft defaults as probe. */
function parseListModelsResult(raw: unknown): RunnerListModelsResult {
  if (!isRecord(raw)) {
    throw new Error("Invalid API response: list-models response must be an object");
  }
  const out: RunnerListModelsResult = {
    ok: typeof raw.ok === "boolean" ? raw.ok : false,
    runner: typeof raw.runner === "string" ? raw.runner : "",
  };
  if (typeof raw.binary_path === "string") out.binary_path = raw.binary_path;
  if (typeof raw.error === "string") out.error = raw.error;
  if (Array.isArray(raw.models)) {
    out.models = raw.models
      .filter((m): m is Record<string, unknown> => isRecord(m))
      .map((m) => ({
        id: typeof m.id === "string" ? m.id : "",
        label: typeof m.label === "string" ? m.label : "",
      }));
  }
  return out;
}
