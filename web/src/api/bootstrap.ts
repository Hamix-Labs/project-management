/**
 * Aggregate cold-start payload from `GET /v1/bootstrap`.
 *
 * The endpoint is an *optimization hint*: the server composes the five
 * existing reads (settings, root tasks page, stats, projects page,
 * drafts head) into one round trip so the SPA can seed its TanStack
 * Query cache without fanning out. Clients MUST tolerate the endpoint
 * being absent (older servers, stripped builds) and fall back to the
 * per-endpoint hooks unchanged.
 *
 * Wire-format guarantees:
 * - `settings` mirrors `GET /settings` exactly.
 * - `tasks` mirrors the `GET /tasks?limit=20&offset=0` envelope.
 * - `stats` mirrors `GET /tasks/stats`.
 * - `projects` mirrors `GET /projects?limit=100`.
 * - `drafts` mirrors `GET /task-drafts?limit=50` (i.e. `{ drafts: [...] }`).
 */
import { fetchWithTimeout, apiErrorFromResponse } from "./shared";
import { parseProjectListResponse } from "./projects";
import {
  parseTaskListResponse,
  parseTaskStatsResponse,
} from "./parseTaskApi";
import { parseTaskDraftSummaryList } from "./parseTaskApi";
import { parseAppSettings, type AppSettings } from "./settings";
import type {
  TaskListResponse,
  TaskStatsResponse,
} from "@/types/task";
import type { ProjectListResponse } from "@/types/project";
import type { TaskDraftSummary } from "@/types/task";

export type Bootstrap = {
  settings: AppSettings;
  tasks: TaskListResponse;
  stats: TaskStatsResponse;
  projects: ProjectListResponse;
  drafts: TaskDraftSummary[];
};

function isRecord(value: unknown): value is Record<string, unknown> {
  return value !== null && typeof value === "object" && !Array.isArray(value);
}

function parseBootstrap(value: unknown): Bootstrap {
  if (!isRecord(value)) {
    throw new Error("Invalid API response: bootstrap must be an object");
  }
  return {
    settings: parseAppSettings(value.settings),
    tasks: parseTaskListResponse(value.tasks),
    stats: parseTaskStatsResponse(value.stats),
    projects: parseProjectListResponse(value.projects),
    drafts: parseTaskDraftSummaryList(value.drafts),
  };
}

/**
 * Returns `null` when the endpoint is unavailable (404 or 405 from a
 * server that has not been updated). Network and 5xx errors still
 * throw so callers can surface them — the bootstrap hook's contract is
 * "fast path when present, transparent fallback when missing".
 */
export async function fetchBootstrap(
  options?: { signal?: AbortSignal },
): Promise<Bootstrap | null> {
  const res = await fetchWithTimeout("/v1/bootstrap", {
    headers: { Accept: "application/json" },
    signal: options?.signal,
  });
  if (res.status === 404 || res.status === 405) {
    return null;
  }
  if (!res.ok) {
    throw await apiErrorFromResponse(res);
  }
  const raw: unknown = await res.json();
  return parseBootstrap(raw);
}
