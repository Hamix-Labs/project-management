import {
  TaskDetailFolderGitGlyph,
  TaskDetailFolderGlyph,
  TaskDetailGitBranchGlyph,
} from "../layout/TaskDetailActionGlyphs";
import { buildGitContextItems, type GitContextFields } from "./commitDisplay";

type Props = {
  context: GitContextFields;
};

function contextIcon(label: string) {
  if (label === "Branch") {
    return <TaskDetailGitBranchGlyph className="task-commits-context-icon" />;
  }
  if (label === "Worktree") {
    return <TaskDetailFolderGlyph className="task-commits-context-icon" />;
  }
  if (label === "Repo") {
    return <TaskDetailFolderGitGlyph className="task-commits-context-icon" />;
  }
  return null;
}

export function GitContextMeta({ context }: Props) {
  const items = buildGitContextItems(context);
  if (items.length === 0) {
    return null;
  }

  return (
    <div className="task-commits-context" data-testid="task-commits-context">
      <dl className="task-commits-context-list">
        {items.map((item) => (
          <div key={item.label} className="task-commits-context-item">
            <dt className="task-commits-context-label">
              {contextIcon(item.label)}
              <span>{item.label}</span>
            </dt>
            <dd className="task-commits-context-value" title={item.title}>
              {item.value}
            </dd>
          </div>
        ))}
      </dl>
    </div>
  );
}
