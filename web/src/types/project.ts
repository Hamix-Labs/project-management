export type ProjectStatus = "active" | "archived";

export type Project = {
  id: string;
  name: string;
  description: string;
  status: ProjectStatus;
  context_summary: string;
  repository_id?: string;
  is_default: boolean;
  created_at: string;
  updated_at: string;
};

export type ProjectContextItem = {
  id: string;
  project_id: string;
  tag: string;
  title: string;
  description: string;
  body: string;
  source_task_id?: string;
  source_cycle_id?: string;
  created_by: "user" | "agent";
  pinned: boolean;
  created_at: string;
  updated_at: string;
};

export type ProjectListResponse = {
  projects: Project[];
  limit: number;
};

export type ProjectContextListResponse = {
  items: ProjectContextItem[];
  /** Always empty; retained for API wire compatibility. */
  edges: [];
  limit: number;
};

export const PROJECT_STATUSES: ProjectStatus[] = ["active", "archived"];
