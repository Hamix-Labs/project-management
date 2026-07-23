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

function CheckGlyph() {
  return (
    <svg
      width="14"
      height="14"
      viewBox="0 0 16 16"
      fill="none"
      stroke="currentColor"
      strokeWidth="2.5"
      strokeLinecap="round"
      strokeLinejoin="round"
      aria-hidden="true"
    >
      <path d="M3.5 8.5L6.5 11.5 12.5 4.5" />
    </svg>
  );
}

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
        Were any of these{" "}
        <span className="task-polish-dialog__not">not</span> done correctly?
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
          const flagged = flaggedIds.has(c.id);
          return (
            <li key={c.id}>
              <label
                htmlFor={inputId}
                className={
                  flagged
                    ? "task-polish-dialog__criteria-card task-polish-dialog__criteria-card--flagged"
                    : "task-polish-dialog__criteria-card"
                }
              >
                <span
                  className={
                    flagged
                      ? "task-polish-dialog__check task-polish-dialog__check--on"
                      : "task-polish-dialog__check"
                  }
                  aria-hidden="true"
                >
                  {flagged ? <CheckGlyph /> : null}
                </span>
                <input
                  id={inputId}
                  type="checkbox"
                  className="visually-hidden"
                  checked={flagged}
                  onChange={() => onToggle(c.id)}
                />
                <span className="task-polish-dialog__criteria-text">{c.text}</span>
              </label>
            </li>
          );
        })}
      </ul>
    </fieldset>
  );
}
