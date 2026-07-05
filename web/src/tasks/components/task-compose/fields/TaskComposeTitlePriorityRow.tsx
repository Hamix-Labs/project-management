import type { PriorityChoice } from "@/types";
import { FieldLabel } from "@/shared/FieldLabel";
import { PrioritySelect } from "./PrioritySelect";

type Props = {
  idsPrefix: string;
  title: string;
  priority: PriorityChoice;
  disabled: boolean;
  onTitleChange: (v: string) => void;
  onPriorityChange: (p: PriorityChoice) => void;
  /** Wider title column for modal essentials grid. */
  layout?: "default" | "modalEssentials";
};

export function TaskComposeTitlePriorityRow({
  idsPrefix,
  title,
  priority,
  disabled,
  onTitleChange,
  onPriorityChange,
  layout = "default",
}: Props) {
  const titleId = `${idsPrefix}-title`;
  const priorityId = `${idsPrefix}-priority`;

  return (
    <div
      className={[
        "task-create-title-row",
        layout === "modalEssentials" ? "task-create-title-row--modal-essentials" : "",
      ]
        .filter(Boolean)
        .join(" ")}
    >
      <div className="field grow">
        <FieldLabel htmlFor={titleId} requirement="required">
          Title
        </FieldLabel>
        <input
          id={titleId}
          className="task-create-title-input"
          value={title}
          onChange={(ev) => onTitleChange(ev.target.value)}
          placeholder="What should get done?"
          required
          aria-required="true"
          disabled={disabled}
        />
      </div>
      <PrioritySelect
        id={priorityId}
        value={priority}
        compact
        onChange={onPriorityChange}
      />
    </div>
  );
}
