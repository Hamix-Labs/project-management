/** Compose route helpers (ADR-0100). */

export function tasksNewPath(opts?: {
  project?: string;
  draft?: string;
  repository?: string;
  worktree?: string;
  lockGit?: boolean;
  lockProject?: boolean;
}): string {
  const qs = new URLSearchParams();
  if (opts?.project) qs.set("project", opts.project);
  if (opts?.draft) qs.set("draft", opts.draft);
  if (opts?.repository) qs.set("repository", opts.repository);
  if (opts?.worktree) qs.set("worktree", opts.worktree);
  if (opts?.lockGit) qs.set("lock_git", "1");
  if (opts?.lockProject) qs.set("lock_project", "1");
  const q = qs.toString();
  return q ? `/tasks/new?${q}` : "/tasks/new";
}

export function taskEditPath(taskId: string): string {
  return `/tasks/${encodeURIComponent(taskId)}/edit`;
}

export function templatesNewPath(): string {
  return "/templates/new";
}

export function templateEditPath(templateId: string): string {
  return `/templates/${encodeURIComponent(templateId)}/edit`;
}
