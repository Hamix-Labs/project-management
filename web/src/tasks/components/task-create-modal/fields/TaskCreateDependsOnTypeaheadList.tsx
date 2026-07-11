import type { Task } from "@/types";
import { shortId } from "./taskCreateDependsOnUtils";

type Props = {
  listboxId: string;
  typeaheadResults: Task[];
  projectTaskCount: number;
  onSelect: (id: string) => void;
};

export function TaskCreateDependsOnTypeaheadList({
  listboxId,
  typeaheadResults,
  projectTaskCount,
  onSelect,
}: Props) {
  return (
    <ul
      id={listboxId}
      role="listbox"
      className="task-create-deps__list"
      aria-label="Matching tasks"
    >
      {typeaheadResults.map((t) => (
        <li key={t.id} role="option" aria-selected="false">
          <button
            type="button"
            className="task-create-deps__option"
            // `mousedown` (not `click`) so the action lands before
            // the input fires its `blur`, otherwise the deferred
            // close races the click and swallows it.
            onMouseDown={(e) => {
              e.preventDefault();
              onSelect(t.id);
            }}
          >
            <span className="task-create-deps__option-title">
              {t.title || "(untitled task)"}
            </span>
            <span className="task-create-deps__option-meta">{shortId(t.id)}</span>
          </button>
        </li>
      ))}
      {typeaheadResults.length === 0 ? (
        <li className="task-create-deps__option task-create-deps__option--empty">
          {projectTaskCount === 0
            ? "No tasks exist in this project yet."
            : "No tasks match."}
        </li>
      ) : null}
    </ul>
  );
}
