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

export type ProjectListResponse = {
  projects: Project[];
  limit: number;
};

export const PROJECT_STATUSES: ProjectStatus[] = ["active", "archived"];
