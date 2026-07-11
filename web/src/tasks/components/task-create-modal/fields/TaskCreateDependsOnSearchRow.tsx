import type { KeyboardEvent } from "react";

type Props = {
  inputId: string;
  listboxId: string;
  hasProject: boolean;
  listOpen: boolean;
  query: string;
  inputDisabled: boolean;
  projectTaskCount: number;
  onQueryChange: (value: string) => void;
  onFocus: () => void;
  onBlur: () => void;
  onKeyDown: (e: KeyboardEvent<HTMLInputElement>) => void;
  onBrowseOpen: () => void;
};

export function TaskCreateDependsOnSearchRow({
  inputId,
  listboxId,
  hasProject,
  listOpen,
  query,
  inputDisabled,
  projectTaskCount,
  onQueryChange,
  onFocus,
  onBlur,
  onKeyDown,
  onBrowseOpen,
}: Props) {
  return (
    <div className="task-create-deps__row">
      <input
        id={inputId}
        type="text"
        className="input task-create-deps__search"
        autoComplete="off"
        role="combobox"
        aria-expanded={listOpen && hasProject}
        aria-controls={listboxId}
        aria-autocomplete="list"
        placeholder={
          hasProject ? "Search tasks by name…" : "Pick a project first"
        }
        disabled={inputDisabled}
        value={query}
        onChange={(e) => onQueryChange(e.target.value)}
        onFocus={onFocus}
        onBlur={onBlur}
        onKeyDown={onKeyDown}
      />
      <button
        type="button"
        className="secondary task-create-deps__browse-btn"
        onClick={onBrowseOpen}
        disabled={inputDisabled || projectTaskCount === 0}
      >
        Browse
      </button>
    </div>
  );
}
