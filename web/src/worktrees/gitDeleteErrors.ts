import { ApiError } from "@/api";

export type GitDeleteTarget =
  | { kind: "repository"; id: string; label: string; repositoryId: string }
  | {
      kind: "worktree";
      mode: "unregister" | "remove_from_disk";
      id: string;
      label: string;
      repositoryId: string;
    };

export function gitDeleteErrorMessage(err: unknown): string {
  if (!(err instanceof ApiError)) {
    return err instanceof Error ? err.message : "Delete failed";
  }
  if (err.code === "branch_checked_out") {
    return "This branch is checked out in another worktree. Switch that worktree to a different branch first.";
  }
  if (err.code === "has_running_task") {
    return err.message || "A task is still running against this git resource.";
  }
  if (err.code === "path_exists" && err.message.includes("uncommitted changes")) {
    return "This worktree has uncommitted changes. Enable force remove to delete anyway.";
  }
  return err.message;
}

export function gitDeleteBlocked(err: unknown): boolean {
  return err instanceof ApiError && err.code === "has_running_task";
}

export function gitDeleteNeedsForce(err: unknown): boolean {
  return (
    err instanceof ApiError &&
    err.code === "path_exists" &&
    err.message.includes("uncommitted changes")
  );
}
