type Props = {
  headingId: string;
  newCriteria: string[];
  draft: string;
  disabled: boolean;
  onDraftChange: (value: string) => void;
  onAdd: () => void;
  onRemove: (index: number) => void;
};

function PlusGlyph() {
  return (
    <svg
      width="16"
      height="16"
      viewBox="0 0 16 16"
      fill="none"
      stroke="currentColor"
      strokeWidth="1.75"
      strokeLinecap="round"
      aria-hidden="true"
    >
      <path d="M8 3.25v9.5M3.25 8h9.5" />
    </svg>
  );
}

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

function CloseGlyph() {
  return (
    <svg
      width="16"
      height="16"
      viewBox="0 0 16 16"
      fill="none"
      stroke="currentColor"
      strokeWidth="1.5"
      strokeLinecap="round"
      aria-hidden="true"
    >
      <path d="M4 4l8 8M12 4l-8 8" />
    </svg>
  );
}

export function TaskPolishAddCriteria({
  headingId,
  newCriteria,
  draft,
  disabled,
  onDraftChange,
  onAdd,
  onRemove,
}: Props) {
  return (
    <div className="task-polish-dialog__add-criteria">
      <h3 id={headingId} className="task-polish-dialog__section-title">
        Add criteria
      </h3>
      <p className="task-polish-dialog__section-hint">
        Optional new requirements for this polish attempt.
      </p>
      <div className="task-polish-dialog__add-row">
        <input
          type="text"
          value={draft}
          disabled={disabled}
          placeholder="New criterion…"
          aria-labelledby={headingId}
          onChange={(e) => onDraftChange(e.target.value)}
          onKeyDown={(e) => {
            if (e.key === "Enter" && !e.nativeEvent.isComposing) {
              e.preventDefault();
              onAdd();
            }
          }}
        />
        <button
          type="button"
          className="secondary task-polish-dialog__add-btn"
          disabled={disabled || !draft.trim()}
          onClick={onAdd}
        >
          <PlusGlyph />
          Add
        </button>
      </div>
      {newCriteria.length > 0 ? (
        <ul
          className="task-polish-dialog__new-list"
          aria-labelledby={headingId}
        >
          {newCriteria.map((text, index) => (
            <li key={`${index}-${text}`} className="task-polish-dialog__new-chip">
              <span
                className="task-polish-dialog__check task-polish-dialog__check--on"
                aria-hidden="true"
              >
                <CheckGlyph />
              </span>
              <span className="task-polish-dialog__criteria-text">{text}</span>
              <button
                type="button"
                className="task-polish-dialog__new-remove"
                disabled={disabled}
                aria-label={`Remove criterion ${text}`}
                onClick={() => onRemove(index)}
              >
                <CloseGlyph />
              </button>
            </li>
          ))}
        </ul>
      ) : null}
    </div>
  );
}
