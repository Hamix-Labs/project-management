import { TaskDetailExternalLinkGlyph } from "./TaskDetailActionGlyphs";

type Props = {
  url: string;
};

/** Opens the task's GitHub PR (Open in–matched utility chrome). */
export function ViewPullRequestLink({ url }: Props) {
  const href = url.trim();
  if (href === "") {
    return null;
  }
  return (
    <a
      className="btn-utility task-detail-view-pr"
      href={href}
      target="_blank"
      rel="noopener noreferrer"
      aria-label="View pull request"
      data-testid="task-detail-view-pr"
    >
      <TaskDetailExternalLinkGlyph className="task-detail-action-glyph" />
      View PR
    </a>
  );
}
