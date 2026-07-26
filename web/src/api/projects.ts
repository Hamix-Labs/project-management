import type {
  Project,
  ProjectListResponse,
  ProjectStatus,
} from "@/types";
import { fetchWithTimeout, jsonHeaders, apiErrorFromResponse } from "./shared";
import {
  isRecord,
  parseBooleanField,
  parseFiniteNumber,
  parseISO8601Required,
  parseNonEmptyString,
  parseOptionalNonEmptyId,
  parseString,
} from "./parseTaskApiCore";
import { assertListIntQuery, assertTaskPathId } from "./taskRequestBounds";

const PROJECT_STATUSES = ["active", "archived"] as const;

function parseProjectStatus(value: unknown): ProjectStatus {
  if (
    typeof value !== "string" ||
    !(PROJECT_STATUSES as readonly string[]).includes(value)
  ) {
    throw new Error("Invalid API response: project status must be active or archived");
  }
  return value as ProjectStatus;
}

export function parseProject(value: unknown): Project {
  if (!isRecord(value)) {
    throw new Error("Invalid API response: project must be an object");
  }
  return {
    id: parseNonEmptyString(value.id, "id"),
    name: parseString(value.name, "name"),
    description: parseString(value.description, "description"),
    status: parseProjectStatus(value.status),
    repository_id: parseOptionalNonEmptyId(value.repository_id, "repository_id"),
    is_default: parseBooleanField(value.is_default, "is_default"),
    created_at: parseISO8601Required(value.created_at, "created_at"),
    updated_at: parseISO8601Required(value.updated_at, "updated_at"),
  };
}

export function parseProjectListResponse(value: unknown): ProjectListResponse {
  if (!isRecord(value)) {
    throw new Error("Invalid API response: project list must be an object");
  }
  const raw = value.projects;
  if (!Array.isArray(raw)) {
    throw new Error("Invalid API response: projects must be an array");
  }
  return {
    projects: raw.map(parseProject),
    limit: parseFiniteNumber(value.limit, "limit"),
  };
}

export async function listProjects(options?: {
  signal?: AbortSignal;
  limit?: number;
  includeArchived?: boolean;
}): Promise<ProjectListResponse> {
  const q = new URLSearchParams({
    limit:
      options?.limit === undefined
        ? "50"
        : assertListIntQuery("limit", options.limit, 0, 100),
  });
  if (options?.includeArchived) q.set("include_archived", "true");
  const res = await fetchWithTimeout(`/projects?${q}`, {
    headers: { Accept: "application/json" },
    signal: options?.signal,
  });
  if (!res.ok) throw await apiErrorFromResponse(res);
  return parseProjectListResponse((await res.json()) as unknown);
}

export async function getProject(
  id: string,
  options?: { signal?: AbortSignal },
): Promise<Project> {
  const projectID = assertTaskPathId(id, "project id");
  const res = await fetchWithTimeout(`/projects/${encodeURIComponent(projectID)}`, {
    headers: { Accept: "application/json" },
    signal: options?.signal,
  });
  if (!res.ok) throw await apiErrorFromResponse(res);
  return parseProject((await res.json()) as unknown);
}

export async function createProject(input: {
  name: string;
  id?: string;
  description?: string;
  repository_id: string;
}): Promise<Project> {
  const res = await fetchWithTimeout("/projects", {
    method: "POST",
    headers: jsonHeaders,
    body: JSON.stringify(input),
  });
  if (!res.ok) throw await apiErrorFromResponse(res);
  return parseProject((await res.json()) as unknown);
}

export async function patchProject(
  id: string,
  input: {
    name?: string;
    description?: string;
    status?: ProjectStatus;
  },
): Promise<Project> {
  const projectID = assertTaskPathId(id, "project id");
  const res = await fetchWithTimeout(`/projects/${encodeURIComponent(projectID)}`, {
    method: "PATCH",
    headers: jsonHeaders,
    body: JSON.stringify(input),
  });
  if (!res.ok) throw await apiErrorFromResponse(res);
  return parseProject((await res.json()) as unknown);
}

export async function deleteProject(id: string): Promise<void> {
  const projectID = assertTaskPathId(id, "project id");
  const res = await fetchWithTimeout(`/projects/${encodeURIComponent(projectID)}`, {
    method: "DELETE",
    headers: { Accept: "application/json" },
  });
  if (!res.ok) throw await apiErrorFromResponse(res);
}
