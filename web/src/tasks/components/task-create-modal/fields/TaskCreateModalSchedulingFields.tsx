import { FieldLabel } from "@/shared/FieldLabel";
import { TaskCreateDependsOnPicker } from "./TaskCreateDependsOnPicker";

type Props = {
  disabled: boolean;
  tagsCsv: string;
  milestone: string;
  /**
   * Project the new task is scoped to. Forwarded to the dependency
   * picker so it filters task lookups by `project_id`. Empty string
   * means "no project bound" — the picker reads as disabled chrome.
   */
  projectId: string;
  dependsOn: string[];
  onTagsCsvChange: (value: string) => void;
  onMilestoneChange: (value: string) => void;
  onDependsOnChange: (value: string[]) => void;
  /** When false, hides the tags CSV field (launch gate). */
  showTags?: boolean;
  /** When false, hides the milestone field (launch gate). */
  showMilestone?: boolean;
  /** When false, hides the depends-on field (detail page owns dependency edits). */
  showDependsOn?: boolean;
  /** When true, depends-on picker is read-only while tags/milestone stay editable. */
  dependsOnDisabled?: boolean;
};

export function TaskCreateModalSchedulingFields({
  disabled,
  tagsCsv,
  milestone,
  projectId,
  dependsOn,
  onTagsCsvChange,
  onMilestoneChange,
  onDependsOnChange,
  showTags = true,
  showMilestone = true,
  showDependsOn = true,
  dependsOnDisabled = false,
}: Props) {
  const showAny = showTags || showMilestone || showDependsOn;
  if (!showAny) {
    return null;
  }

  const legend =
    showTags && !showMilestone && !showDependsOn
      ? "Tags"
      : showTags
        ? "Tags & dependencies"
        : "Dependencies";

  return (
    <fieldset className="task-create-scheduling" disabled={disabled}>
      <legend className="task-create-scheduling__legend">{legend}</legend>
      <div className="task-create-scheduling__grid">
        {showTags ? (
          <div className="task-create-scheduling__field">
            <FieldLabel htmlFor="create-tags">Tags</FieldLabel>
            <input
              id="create-tags"
              className="input"
              value={tagsCsv}
              onChange={(e) => onTagsCsvChange(e.target.value)}
              placeholder="e.g. backend, api"
              aria-describedby="create-tags-hint"
            />
            <p id="create-tags-hint" className="hint">
              Lowercase letters, numbers, and . _ - (capitals are normalized on
              save).
            </p>
          </div>
        ) : null}
        {showMilestone ? (
          <div className="task-create-scheduling__field">
            <FieldLabel htmlFor="create-milestone">Milestone</FieldLabel>
            <input
              id="create-milestone"
              className="input"
              value={milestone}
              onChange={(e) => onMilestoneChange(e.target.value)}
              placeholder="e.g. M1 — auth"
            />
          </div>
        ) : null}
        {showDependsOn ? (
          <div className="task-create-scheduling__field task-create-scheduling__field--full">
            <TaskCreateDependsOnPicker
              projectId={projectId}
              selected={dependsOn}
              onChange={onDependsOnChange}
              disabled={disabled || dependsOnDisabled}
            />
          </div>
        ) : null}
      </div>
    </fieldset>
  );
}
