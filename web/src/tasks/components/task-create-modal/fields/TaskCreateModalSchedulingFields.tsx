import { FieldLabel } from "@/shared/FieldLabel";
import { AgentListIcon, AgentTagIcon } from "./TaskCreateAgentIcons";
import { TaskCreateConfigSectionHeader } from "./TaskCreateConfigSectionHeader";
import { TaskCreateDependsOnPicker } from "./TaskCreateDependsOnPicker";
import { TaskCreateTagsPillsField } from "./TaskCreateTagsPillsField";

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
  /** When true, use the agent-config section chrome (create modal advanced body). */
  configChrome?: boolean;
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
  configChrome = false,
}: Props) {
  const showAny = showTags || showMilestone || showDependsOn;
  if (!showAny) {
    return null;
  }

  const showDepsBlock = showMilestone || showDependsOn;

  if (configChrome) {
    return (
      <div className="task-create-scheduling task-create-scheduling--config">
        {showTags ? (
          <div className="task-create-config-section">
            <TaskCreateConfigSectionHeader
              id="task-create-tags-heading"
              title="Tags"
              icon={<AgentTagIcon />}
            />
            <TaskCreateTagsPillsField
              id="create-tags"
              disabled={disabled}
              tagsCsv={tagsCsv}
              onTagsCsvChange={onTagsCsvChange}
            />
          </div>
        ) : null}

        {showTags && showDepsBlock ? (
          <div className="task-create-advanced__divider" role="separator" />
        ) : null}

        {showDepsBlock ? (
          <div className="task-create-config-section">
            <TaskCreateConfigSectionHeader
              id="task-create-deps-heading"
              title={
                showMilestone && showDependsOn
                  ? "Dependencies"
                  : showMilestone
                    ? "Milestone"
                    : "Dependencies"
              }
              icon={<AgentListIcon />}
            />
            <div className="task-create-scheduling__grid">
              {showMilestone ? (
                <div className="task-create-scheduling__field">
                  <FieldLabel htmlFor="create-milestone">Milestone</FieldLabel>
                  <input
                    id="create-milestone"
                    className="input"
                    value={milestone}
                    disabled={disabled}
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
          </div>
        ) : null}
      </div>
    );
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
            <TaskCreateTagsPillsField
              id="create-tags"
              disabled={disabled}
              tagsCsv={tagsCsv}
              onTagsCsvChange={onTagsCsvChange}
            />
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
