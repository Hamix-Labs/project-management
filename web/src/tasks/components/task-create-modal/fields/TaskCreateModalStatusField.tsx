import { CLIENT_WRITABLE_STATUSES, type Status } from "@/types";
import { FieldLabel } from "@/shared/FieldLabel";
import { statusListLabel } from "@/lib/taskStatusDisplay";

type Props = {
  id?: string;
  status: Status;
  disabled: boolean;
  onChange: (status: Status) => void;
};

export function TaskCreateModalStatusField({
  id = "task-edit-status",
  status,
  disabled,
  onChange,
}: Props) {
  const options = CLIENT_WRITABLE_STATUSES.includes(status)
    ? CLIENT_WRITABLE_STATUSES
    : [...CLIENT_WRITABLE_STATUSES, status];

  return (
    <div className="field grow">
      <FieldLabel htmlFor={id} requirement="required">
        Status
      </FieldLabel>
      <select
        aria-required="true"
        id={id}
        value={status}
        disabled={disabled}
        onChange={(ev) => onChange(ev.target.value as Status)}
      >
        {options.map((s) => (
          <option key={s} value={s}>
            {statusListLabel(s)}
          </option>
        ))}
      </select>
    </div>
  );
}
