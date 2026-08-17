import { AgentTagIcon } from "../../task-create-modal/fields/TaskCreateAgentIcons";
import { TaskCreateTagsPillsField } from "../../task-create-modal/fields/TaskCreateTagsPillsField";
import { ComposeRailSectionTitle } from "../TaskComposeBriefCard";

type Props = {
  tagsCsv: string;
  disabled: boolean;
  onTagsCsvChange: (value: string) => void;
};

export function TaskComposeTagsCard({
  tagsCsv,
  disabled,
  onTagsCsvChange,
}: Props) {
  return (
    <section className="compose-handoff__section compose-tags">
      <ComposeRailSectionTitle icon={<AgentTagIcon />}>
        Tags
      </ComposeRailSectionTitle>
      <div className="compose-tags__field">
        <TaskCreateTagsPillsField
          id="compose-tags"
          disabled={disabled}
          tagsCsv={tagsCsv}
          onTagsCsvChange={onTagsCsvChange}
        />
      </div>
    </section>
  );
}
