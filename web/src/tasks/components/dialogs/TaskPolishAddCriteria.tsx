type Props = {
  headingId: string;
  newCriteria: string[];
  draft: string;
  disabled: boolean;
  onDraftChange: (value: string) => void;
  onAdd: () => void;
  onRemove: (index: number) => void;
};

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
      {newCriteria.length > 0 ? (
        <ul
          className="task-polish-dialog__new-list"
          aria-labelledby={headingId}
        >
          {newCriteria.map((text, index) => (
            <li key={`${index}-${text}`} className="task-polish-dialog__new-item">
              <span>{text}</span>
              <button
                type="button"
                className="secondary task-polish-dialog__new-remove"
                disabled={disabled}
                onClick={() => onRemove(index)}
              >
                Remove
              </button>
            </li>
          ))}
        </ul>
      ) : null}
      <div className="task-polish-dialog__add-row">
        <input
          type="text"
          value={draft}
          disabled={disabled}
          placeholder="New criterion…"
          aria-labelledby={headingId}
          onChange={(e) => onDraftChange(e.target.value)}
          onKeyDown={(e) => {
            if (e.key === "Enter") {
              e.preventDefault();
              onAdd();
            }
          }}
        />
        <button
          type="button"
          className="secondary"
          disabled={disabled || !draft.trim()}
          onClick={onAdd}
        >
          Add
        </button>
      </div>
    </div>
  );
}
