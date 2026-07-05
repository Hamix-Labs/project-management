import type {
  TaskComposePayload,
  TaskTemplateDetail,
  TaskTemplateListParams,
  TaskTemplateSummary,
} from "@/types";
import {
  parseTaskTemplateDetail,
  parseTaskTemplateInstantiateResponse,
  parseTaskTemplateSummaryList,
} from "./parseTaskApiTemplates";
import { fetchNamedEntityJson, fetchNamedEntityVoid } from "./namedEntityClient";
import { apiErrorFromResponse, fetchWithTimeout, jsonHeaders } from "./shared";
import {
  assertInstantiateTemplateItems,
  assertListIntQuery,
  assertOptionalTaskPathId,
  assertTaskPathId,
  type TaskTemplateInstantiateItem,
} from "./taskRequestBounds";

export type { TaskTemplateInstantiateItem };
export { maxTemplateInstantiateCountPerItem } from "./taskRequestBounds";

export async function listTaskTemplates(
  options?: TaskTemplateListParams & { signal?: AbortSignal },
): Promise<TaskTemplateSummary[]> {
  const limit = assertListIntQuery("limit", options?.limit ?? 50, 0, 100);
  const params = new URLSearchParams({ limit: String(limit) });
  const q = options?.q?.trim();
  if (q) params.set("q", q);
  if (options?.sort) params.set("sort", options.sort);
  if (options?.order) params.set("order", options.order);
  const tag = options?.tag?.trim();
  if (tag) params.set("tag", tag);
  const res = await fetchWithTimeout(`/task-templates?${params.toString()}`, {
    headers: { Accept: "application/json" },
    signal: options?.signal,
  });
  if (!res.ok) throw await apiErrorFromResponse(res);
  const raw: unknown = await res.json();
  return parseTaskTemplateSummaryList(raw);
}

export async function saveTaskTemplate(input: {
  id?: string;
  name?: string;
  payload: TaskComposePayload;
}): Promise<TaskTemplateSummary> {
  const sid = assertOptionalTaskPathId(input.id, "id");
  const res = await fetchWithTimeout("/task-templates", {
    method: "POST",
    headers: jsonHeaders,
    body: JSON.stringify({
      payload: input.payload,
      ...(input.name !== undefined ? { name: input.name } : {}),
      ...(sid !== undefined ? { id: sid } : {}),
    }),
  });
  if (!res.ok) throw await apiErrorFromResponse(res);
  const raw: unknown = await res.json();
  const list = parseTaskTemplateSummaryList({ templates: [raw] });
  return list[0]!;
}

export async function getTaskTemplate(
  id: string,
  options?: { signal?: AbortSignal },
): Promise<TaskTemplateDetail> {
  const tid = assertTaskPathId(id, "template id");
  return fetchNamedEntityJson(
    `/task-templates/${encodeURIComponent(tid)}`,
    { headers: { Accept: "application/json" }, signal: options?.signal },
    parseTaskTemplateDetail,
  );
}

export async function patchTaskTemplate(
  id: string,
  patch: { name?: string; payload?: TaskComposePayload },
): Promise<TaskTemplateDetail> {
  const tid = assertTaskPathId(id, "template id");
  return fetchNamedEntityJson(
    `/task-templates/${encodeURIComponent(tid)}`,
    { method: "PATCH", headers: jsonHeaders, body: JSON.stringify(patch) },
    parseTaskTemplateDetail,
  );
}

export async function deleteTaskTemplate(id: string): Promise<void> {
  const tid = assertTaskPathId(id, "template id");
  await fetchNamedEntityVoid(`/task-templates/${encodeURIComponent(tid)}`, {
    method: "DELETE",
  });
}

export async function instantiateTaskTemplates(
  items: TaskTemplateInstantiateItem[],
): Promise<{ tasks: import("@/types").Task[]; errors: { template_id: string; error: string }[] }> {
  const normalized = assertInstantiateTemplateItems(items);
  return fetchNamedEntityJson(
    "/task-templates/instantiate",
    { method: "POST", headers: jsonHeaders, body: JSON.stringify({ items: normalized }) },
    parseTaskTemplateInstantiateResponse,
  );
}
