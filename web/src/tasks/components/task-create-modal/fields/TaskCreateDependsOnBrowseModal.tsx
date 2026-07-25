import { Modal } from "@/shared/Modal";
import type { Task } from "@/types";
import { taskDisplayRef } from "@/lib/taskShortId";

type Props = {
  browseTitleId: string;
  browseQuery: string;
  browseResults: Task[];
  selectedSet: Set<string>;
  selectedCount: number;
  disabled: boolean;
  onBrowseQueryChange: (value: string) => void;
  onClose: () => void;
  onToggle: (id: string) => void;
};

export function TaskCreateDependsOnBrowseModal({
  browseTitleId,
  browseQuery,
  browseResults,
  selectedSet,
  selectedCount,
  disabled,
  onBrowseQueryChange,
  onClose,
  onToggle,
}: Props) {
  return (
    <Modal
      onClose={onClose}
      labelledBy={browseTitleId}
      stack="nested"
      lockBodyScroll={false}
    >
      <section className="panel task-create-deps-browse">
        <header className="task-create-deps-browse__header">
          <h3 id={browseTitleId} className="task-create-deps-browse__title">
            Project tasks
          </h3>
          <p className="task-create-deps-browse__lede">
            Toggle the tasks this one should wait for.
          </p>
        </header>
        <div className="task-create-deps-browse__search">
          <input
            type="text"
            className="input"
            placeholder="Search tasks…"
            value={browseQuery}
            onChange={(e) => onBrowseQueryChange(e.target.value)}
            aria-label="Filter project tasks"
            autoFocus
          />
        </div>
        {browseResults.length > 0 ? (
          <ul className="task-create-deps-browse__list">
            {browseResults.map((t) => {
              const checked = selectedSet.has(t.id);
              return (
                <li key={t.id} className="task-create-deps-browse__item">
                  <label className="task-create-deps-browse__row">
                    <input
                      type="checkbox"
                      className="task-create-deps-browse__check"
                      checked={checked}
                      onChange={() => onToggle(t.id)}
                      disabled={disabled}
                    />
                    <span className="task-create-deps-browse__title-cell">
                      <span className="task-create-deps-browse__row-title">
                        {t.title || "(untitled task)"}
                      </span>
                      <span className="task-create-deps-browse__row-meta">
                        {taskDisplayRef(t)} · {t.status}
                      </span>
                    </span>
                  </label>
                </li>
              );
            })}
          </ul>
        ) : (
          <p className="task-create-deps-browse__empty">No tasks match.</p>
        )}
        <footer className="task-create-deps-browse__footer">
          <span className="task-create-deps-browse__count">
            {selectedCount === 1
              ? "1 dependency selected"
              : `${selectedCount} dependencies selected`}
          </span>
          <button
            type="button"
            className="task-create-deps-browse__done"
            onClick={onClose}
          >
            Done
          </button>
        </footer>
      </section>
    </Modal>
  );
}
