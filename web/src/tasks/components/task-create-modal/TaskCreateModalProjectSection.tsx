import type { ReactNode } from "react";
import { TaskCreateModalSection } from "./fields/TaskCreateModalSection";

type Props = {
  projectAssignment: ReactNode;
};

export function TaskCreateModalProjectSection({ projectAssignment }: Props) {
  return (
    <TaskCreateModalSection
      variant="context"
      title="Project"
      lede="Scope this task to a project and attach context the agent can reference."
    >
      {projectAssignment}
    </TaskCreateModalSection>
  );
}
