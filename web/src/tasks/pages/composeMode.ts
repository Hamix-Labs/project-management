export type ComposeMode =
  | { kind: "task-create" }
  | { kind: "task-edit"; taskId: string }
  | { kind: "template-create" }
  | { kind: "template-edit"; templateId: string };

export function composeBackTo(mode: ComposeMode): string {
  if (mode.kind === "template-create" || mode.kind === "template-edit") {
    return "/templates";
  }
  if (mode.kind === "task-edit") {
    return `/tasks/${mode.taskId}`;
  }
  return "/";
}
