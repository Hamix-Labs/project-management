import { assertTaskPathId } from "./taskRequestBounds";
import { apiErrorFromResponse, fetchWithTimeout, jsonHeaders } from "./shared";

/** Resolves the git API root for global or project-scoped routes. */
export function gitApiRoot(projectId?: string): string {
  if (projectId === undefined) {
    return "/git";
  }
  const pid = assertTaskPathId(projectId, "project id");
  return `/projects/${encodeURIComponent(pid)}/git`;
}

export async function gitFetchJson(
  path: string,
  init?: RequestInit,
): Promise<unknown> {
  const res = await fetchWithTimeout(path, init);
  if (!res.ok) {
    throw await apiErrorFromResponse(res);
  }
  return res.json() as Promise<unknown>;
}

export async function gitFetchVoid(path: string, init?: RequestInit): Promise<void> {
  const res = await fetchWithTimeout(path, init);
  if (!res.ok) {
    throw await apiErrorFromResponse(res);
  }
}

export const gitJsonGetInit = (signal?: AbortSignal): RequestInit => ({
  headers: { Accept: "application/json" },
  signal,
});

export const gitJsonPostInit = (body: unknown): RequestInit => ({
  method: "POST",
  headers: jsonHeaders,
  body: JSON.stringify(body),
});

export const gitDeleteInit = (): RequestInit => ({ method: "DELETE" });
