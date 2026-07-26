import type { TaskComposePayload } from "./taskCore";

export type TaskTemplateSummary = {
  id: string;
  name: string;
  created_at: string;
  updated_at: string;
  primary_tag?: string;
  instantiate_count: number;
  is_function?: boolean;
  input_kinds?: Array<"dir" | "file" | "function">;
};

export type TaskTemplateListParams = {
  q?: string;
  limit?: number;
  sort?: "updated_at" | "name" | "instantiate_count";
  order?: "asc" | "desc";
  tag?: string;
};

export type TaskTemplateDetail = TaskTemplateSummary & {
  payload: TaskComposePayload;
};
