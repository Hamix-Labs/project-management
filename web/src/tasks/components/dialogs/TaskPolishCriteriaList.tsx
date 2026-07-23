type PolishCriterionOption = {
  id: string;
  text: string;
};

type Props = {
  criteria: PolishCriterionOption[];
  flaggedIds: Set<string>;
  disabled: boolean;
  headingId: string;
  onToggle: (id: string) => void;
};

export function TaskPolishCriteriaList({
  criteria,
  flaggedIds,
  disabled,
  headingId,
  onToggle,
}: Props) {
  if (criteria.length === 0) return null;

  return (
    <fieldset
      className="task-polish-dialog__criteria"
      disabled={disabled}
    >
      <legend id={headingId} className="task-polish-dialog__section-title">
        Were any of these not done correctly?
      </legend>
      <p className="task-polish-dialog__section-hint">
        Check criteria the agent got wrong. Unchecked criteria stay accepted.
      </p>
      <ul
        className="task-polish-dialog__criteria-list"
        aria-labelledby={headingId}
      >
        {criteria.map((c) => {
          const inputId = `polish-flag-${c.id}`;
          return (
            <li key={c.id} className="task-polish-dialog__criteria-item">
              <input
                id={inputId}
                type="checkbox"
                checked={flaggedIds.has(c.id)}
                onChange={() => onToggle(c.id)}
              />
              <label htmlFor={inputId}>{c.text}</label>
            </li>
          );
        })}
      </ul>
    </fieldset>
  );
}
