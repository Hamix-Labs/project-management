import { promptHasVisibleContent } from "@/lib/promptFormat";
import { TaskDetailCollapsibleSection } from "../layout/TaskDetailCollapsibleSection";

type TaskDetailPromptSectionProps = {
  initialPrompt: string;
  sanitizedInitialPrompt: string;
};

export function TaskDetailPromptSection({
  initialPrompt,
  sanitizedInitialPrompt,
}: TaskDetailPromptSectionProps) {
  if (!promptHasVisibleContent(initialPrompt)) {
    return (
      <div className="task-detail-section task-detail-prompt">
        <h3
          className="task-detail-section-heading"
          id="task-detail-prompt-heading"
        >
          <span>Initial prompt</span>
        </h3>
        <p
          className="muted task-detail-prompt-empty"
          aria-labelledby="task-detail-prompt-heading"
        >
          —
        </p>
      </div>
    );
  }

  return (
    <TaskDetailCollapsibleSection
      as="div"
      className="task-detail-prompt"
      title="Initial prompt"
      headingId="task-detail-prompt-heading"
      defaultOpen={false}
    >
      <div
        className="task-detail-prompt-body"
        dangerouslySetInnerHTML={{ __html: sanitizedInitialPrompt }}
      />
    </TaskDetailCollapsibleSection>
  );
}
