import { shortId } from "./taskCreateDependsOnUtils";

type Props = {
  selected: string[];
  labelLookup: Map<string, string>;
  disabled: boolean;
  onRemove: (id: string) => void;
};

export function TaskCreateDependsOnSelectedChips({
  selected,
  labelLookup,
  disabled,
  onRemove,
}: Props) {
  if (selected.length === 0) return null;

  return (
    <ul className="task-create-deps__chips" aria-label="Selected dependencies">
      {selected.map((id) => (
        <li key={id}>
          <button
            type="button"
            className="task-create-deps__chip"
            onClick={() => onRemove(id)}
            disabled={disabled}
            aria-label={`Remove dependency ${labelLookup.get(id) ?? shortId(id)}`}
          >
            <span className="task-create-deps__chip-title">
              {labelLookup.get(id) ?? shortId(id)}
            </span>
            <span className="task-create-deps__chip-remove" aria-hidden="true">
              ×
            </span>
          </button>
        </li>
      ))}
    </ul>
  );
}
