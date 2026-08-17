import { TaskCreateTagsPillsField } from "../../task-create-modal/fields/TaskCreateTagsPillsField";

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
      <h2 className="compose-handoff__title">Tags</h2>
      <div className="compose-tags__field">
        <TaskCreateTagsPillsField
          id="compose-tags"
          disabled={disabled}
          tagsCsv={tagsCsv}
          onTagsCsvChange={onTagsCsvChange}
          hideHint
          placeholder="backend, api"
        />
      </div>
    </section>
  );
}
