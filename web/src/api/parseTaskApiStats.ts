import {
  CYCLE_STATUSES,
  PHASE_STATUSES,
  PHASES,
  PRIORITIES,
  STATUSES,
  type CycleFailuresListResponse,
  type CycleStatus,
  type PhaseStatus,
  type WritablePhase,
  type Priority,
  type Status,
  type TaskStatsCycles,
  type TaskStatsPhases,
  type TaskStatsRecentFailure,
  type TaskStatsResponse,
  type TaskStatsRunner,
  type TaskStatsRunnerBucket,
} from "@/types";
import {
  isRecord,
  parseBooleanField,
  parseFiniteNumber,
  parseNonEmptyString,
  parseString,
} from "./parseTaskApiCore";

/** Validates GET /tasks/stats JSON. */
export function parseTaskStatsResponse(value: unknown): TaskStatsResponse {
  if (!isRecord(value)) {
    throw new Error("Invalid API response: task stats payload must be an object");
  }
  const byStatusRaw = value.by_status;
  if (!isRecord(byStatusRaw)) {
    throw new Error("Invalid API response: by_status must be an object");
  }
  const byPriorityRaw = value.by_priority;
  if (!isRecord(byPriorityRaw)) {
    throw new Error("Invalid API response: by_priority must be an object");
  }
  const by_status: Partial<Record<Status, number>> = {};
  for (const [key, rawCount] of Object.entries(byStatusRaw)) {
    if (!(STATUSES as readonly string[]).includes(key)) {
      throw new Error(`Invalid API response: by_status.${key} is not a known status`);
    }
    by_status[key as Status] = parseFiniteNumber(rawCount, `by_status.${key}`);
  }
  const by_priority: Partial<Record<Priority, number>> = {};
  for (const [key, rawCount] of Object.entries(byPriorityRaw)) {
    if (!(PRIORITIES as readonly string[]).includes(key)) {
      throw new Error(`Invalid API response: by_priority.${key} is not a known priority`);
    }
    by_priority[key as Priority] = parseFiniteNumber(rawCount, `by_priority.${key}`);
  }
  const cycles = parseTaskStatsCycles(value.cycles);
  const phases = parseTaskStatsPhases(value.phases);
  const runner = parseTaskStatsRunner(value.runner);
  const recent_failures = parseTaskStatsRecentFailures(value.recent_failures);
  return {
    total: parseFiniteNumber(value.total, "total"),
    ready: parseFiniteNumber(value.ready, "ready"),
    critical: parseFiniteNumber(value.critical, "critical"),
    scheduled:
      value.scheduled === undefined
        ? 0
        : parseFiniteNumber(value.scheduled, "scheduled"),
    by_status,
    by_priority,
    cycles,
    phases,
    runner,
    recent_failures,
  };
}

export function parseCycleFailuresListResponse(
  value: unknown,
): CycleFailuresListResponse {
  if (!isRecord(value)) {
    throw new Error(
      "Invalid API response: cycle failures list payload must be an object",
    );
  }
  return {
    total: parseFiniteNumber(value.total, "total"),
    limit: parseFiniteNumber(value.limit, "limit"),
    offset: parseFiniteNumber(value.offset, "offset"),
    sort: parseString(value.sort, "sort"),
    reason_sort_truncated: parseBooleanField(
      value.reason_sort_truncated,
      "reason_sort_truncated",
    ),
    failures: parseTaskStatsRecentFailures(value.failures),
  };
}

function parseTaskStatsCycles(value: unknown): TaskStatsCycles {
  if (!isRecord(value)) {
    throw new Error("Invalid API response: cycles must be an object");
  }
  const byStatusRaw = value.by_status;
  if (!isRecord(byStatusRaw)) {
    throw new Error("Invalid API response: cycles.by_status must be an object");
  }
  const byTriggeredByRaw = value.by_triggered_by;
  if (!isRecord(byTriggeredByRaw)) {
    throw new Error("Invalid API response: cycles.by_triggered_by must be an object");
  }
  const by_status: Partial<Record<CycleStatus, number>> = {};
  for (const [key, rawCount] of Object.entries(byStatusRaw)) {
    if (!(CYCLE_STATUSES as readonly string[]).includes(key)) {
      throw new Error(`Invalid API response: cycles.by_status.${key} is not a known cycle status`);
    }
    by_status[key as CycleStatus] = parseFiniteNumber(rawCount, `cycles.by_status.${key}`);
  }
  const by_triggered_by: Partial<Record<"user" | "agent", number>> = {};
  for (const [key, rawCount] of Object.entries(byTriggeredByRaw)) {
    if (key !== "user" && key !== "agent") {
      throw new Error(
        `Invalid API response: cycles.by_triggered_by.${key} must be "user" or "agent"`,
      );
    }
    by_triggered_by[key] = parseFiniteNumber(rawCount, `cycles.by_triggered_by.${key}`);
  }
  return { by_status, by_triggered_by };
}

function parseTaskStatsPhases(value: unknown): TaskStatsPhases {
  if (!isRecord(value)) {
    throw new Error("Invalid API response: phases must be an object");
  }
  const byPhaseStatusRaw = value.by_phase_status;
  if (!isRecord(byPhaseStatusRaw)) {
    throw new Error("Invalid API response: phases.by_phase_status must be an object");
  }
  const by_phase_status = {} as TaskStatsPhases["by_phase_status"];
  // PHASES is the canonical writable set (`execute`, `verify`). We
  // narrow it to `WritablePhase[]` here because the stats heatmap
  // type is keyed on the writable subset only — legacy phase values
  // are dropped at the parser boundary below.
  const writablePhases = PHASES as readonly WritablePhase[];
  for (const phase of writablePhases) {
    by_phase_status[phase] = {};
  }
  // Legacy phase rows (diagnose / persist) may surface from historical
  // task_cycle_phases data even though the worker no longer writes them.
  // Drop the bucket silently rather than refusing to render the heatmap.
  const writablePhaseKeys = writablePhases as readonly string[];
  for (const [phaseKey, inner] of Object.entries(byPhaseStatusRaw)) {
    if (!writablePhaseKeys.includes(phaseKey)) {
      continue;
    }
    if (!isRecord(inner)) {
      throw new Error(
        `Invalid API response: phases.by_phase_status.${phaseKey} must be an object`,
      );
    }
    const bucket: Partial<Record<PhaseStatus, number>> = {};
    for (const [statusKey, rawCount] of Object.entries(inner)) {
      if (!(PHASE_STATUSES as readonly string[]).includes(statusKey)) {
        throw new Error(
          `Invalid API response: phases.by_phase_status.${phaseKey}.${statusKey} is not a known phase status`,
        );
      }
      bucket[statusKey as PhaseStatus] = parseFiniteNumber(
        rawCount,
        `phases.by_phase_status.${phaseKey}.${statusKey}`,
      );
    }
    by_phase_status[phaseKey as WritablePhase] = bucket;
  }
  return { by_phase_status };
}

function parseTaskStatsRunner(value: unknown): TaskStatsRunner {
  if (!isRecord(value)) {
    throw new Error("Invalid API response: runner must be an object");
  }
  return {
    by_runner: parseRunnerBucketMap(value.by_runner, "runner.by_runner"),
    by_model: parseRunnerBucketMap(value.by_model, "runner.by_model"),
    by_runner_model: parseRunnerBucketMap(value.by_runner_model, "runner.by_runner_model"),
    by_runner_model_resolved:
      value.by_runner_model_resolved === undefined
        ? {}
        : parseRunnerBucketMap(
            value.by_runner_model_resolved,
            "runner.by_runner_model_resolved",
          ),
  };
}

function parseRunnerBucketMap(
  value: unknown,
  field: string,
): Record<string, TaskStatsRunnerBucket> {
  if (!isRecord(value)) {
    throw new Error(`Invalid API response: ${field} must be an object`);
  }
  const out: Record<string, TaskStatsRunnerBucket> = {};
  for (const [key, raw] of Object.entries(value)) {
    out[key] = parseRunnerBucket(raw, `${field}.${key}`);
  }
  return out;
}

function parseRunnerBucket(value: unknown, field: string): TaskStatsRunnerBucket {
  if (!isRecord(value)) {
    throw new Error(`Invalid API response: ${field} must be an object`);
  }
  const byStatusRaw = value.by_status;
  if (!isRecord(byStatusRaw)) {
    throw new Error(`Invalid API response: ${field}.by_status must be an object`);
  }
  const by_status: Partial<Record<CycleStatus, number>> = {};
  for (const [statusKey, rawCount] of Object.entries(byStatusRaw)) {
    if (!(CYCLE_STATUSES as readonly string[]).includes(statusKey)) {
      throw new Error(
        `Invalid API response: ${field}.by_status.${statusKey} is not a known cycle status`,
      );
    }
    by_status[statusKey as CycleStatus] = parseFiniteNumber(
      rawCount,
      `${field}.by_status.${statusKey}`,
    );
  }
  return {
    by_status,
    succeeded: parseFiniteNumber(value.succeeded, `${field}.succeeded`),
    duration_p50_succeeded_seconds: parseFiniteNumber(
      value.duration_p50_succeeded_seconds,
      `${field}.duration_p50_succeeded_seconds`,
    ),
    duration_p95_succeeded_seconds: parseFiniteNumber(
      value.duration_p95_succeeded_seconds,
      `${field}.duration_p95_succeeded_seconds`,
    ),
  };
}

function parseTaskStatsRecentFailures(value: unknown): TaskStatsRecentFailure[] {
  if (!Array.isArray(value)) {
    throw new Error("Invalid API response: recent_failures must be an array");
  }
  return value.map((item, i) => parseTaskStatsRecentFailure(item, i));
}

function parseTaskStatsRecentFailure(
  value: unknown,
  index: number,
): TaskStatsRecentFailure {
  if (!isRecord(value)) {
    throw new Error(`Invalid API response: recent_failures[${index}] must be an object`);
  }
  const status = value.status;
  if (status !== "failed" && status !== "aborted") {
    throw new Error(
      `Invalid API response: recent_failures[${index}].status must be "failed" or "aborted"`,
    );
  }
  return {
    task_id: parseNonEmptyString(value.task_id, `recent_failures[${index}].task_id`),
    event_seq: parseFiniteNumber(value.event_seq, `recent_failures[${index}].event_seq`),
    at: parseNonEmptyString(value.at, `recent_failures[${index}].at`),
    cycle_id: parseNonEmptyString(value.cycle_id, `recent_failures[${index}].cycle_id`),
    attempt_seq: parseFiniteNumber(
      value.attempt_seq,
      `recent_failures[${index}].attempt_seq`,
    ),
    status,
    reason: parseString(value.reason, `recent_failures[${index}].reason`),
  };
}
