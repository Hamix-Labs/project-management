import type { Project } from "@/types";
import { FACTORY_GIT_REPO_ID } from "./git";

export const FACTORY_REPO_DEFAULT_PROJECT_ID =
  "00000000-0000-4000-8000-000000000040";

export function repoDefaultProjectFactory(
  overrides: Partial<Project> = {},
): Project {
  return {
    id: FACTORY_REPO_DEFAULT_PROJECT_ID,
    name: "Default",
    description: "",
    status: "active",
    context_summary: "",
    repository_id: FACTORY_GIT_REPO_ID,
    is_default: true,
    created_at: "2026-06-22T12:00:00Z",
    updated_at: "2026-06-22T12:00:00Z",
    ...overrides,
  };
}

export function projectFactory(overrides: Partial<Project> = {}): Project {
  return {
    id: "11111111-1111-4111-8111-111111111111",
    name: "Custom project",
    description: "",
    status: "active",
    context_summary: "",
    repository_id: FACTORY_GIT_REPO_ID,
    is_default: false,
    created_at: "2026-06-22T12:00:00Z",
    updated_at: "2026-06-22T12:00:00Z",
    ...overrides,
  };
}
