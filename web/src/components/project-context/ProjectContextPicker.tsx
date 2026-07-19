import { useCallback, useMemo, useState } from "react";
import type { ProjectContextItem } from "@/types";
import { Modal } from "@/shared/Modal";
import { ProjectContextListView } from "@/projects/ProjectContextListView";
import {
  mergeProjectContextSelection,
  projectContextShortId,
  selectedProjectContextItems,
} from "@/lib/projectContextRefs";
import { useProjectContext } from "@/projects/hooks";

interface ProjectContextPickerProps {
  projectId: string;
  selectedIds: string[];
  disabled?: boolean;
  /** Shorter copy for the create-task modal. */
  compact?: boolean;
  onChange: (ids: string[]) => void;
}

const EMPTY_CONTEXT_ITEMS: ProjectContextItem[] = [];

/**
 * Compact summary of the project context items currently attached to the
 * task plus an entry point into the full chooser. Selecting a node adds that
 * node only (same as the editor's `#` mention flow).
 *
 * The displayed labels intentionally include a 6-character short id so
 * operators can disambiguate same-titled nodes from different projects
 * without opening every detail panel.
 */
export function ProjectContextPicker({
  projectId,
  selectedIds,
  disabled,
  compact = false,
  onChange,
}: ProjectContextPickerProps) {
  const [chooserOpen, setChooserOpen] = useState(false);

  const contextQuery = useProjectContext(projectId, {
    enabled: Boolean(projectId),
    limit: 100,
    pinnedOnly: false,
  });
  const items = contextQuery.data?.items ?? EMPTY_CONTEXT_ITEMS;
  const selected = useMemo(() => new Set(selectedIds), [selectedIds]);
  const selectedItems = useMemo(
    () => selectedProjectContextItems(items, selectedIds),
    [items, selectedIds],
  );
  const selectedCountLabel = compact
    ? `${selectedIds.length} selected`
    : selectedIds.length === 1
      ? "1 node selected"
      : `${selectedIds.length} nodes selected`;

  const compactSummaryText = useMemo(() => {
    if (!compact) return "";
    if (contextQuery.isPending) return "Loading project context…";
    if (contextQuery.error) return "Project context unavailable.";
    if (items.length === 0) return "This project has no context nodes yet.";
    if (selectedItems.length === 0) return "No context attached yet";
    return selectedItems.map((item) => item.title || "(untitled)").join(", ");
  }, [
    compact,
    contextQuery.error,
    contextQuery.isPending,
    items.length,
    selectedItems,
  ]);

  const handleToggle = useCallback(
    (item: ProjectContextItem) => {
      if (disabled) return;
      if (selected.has(item.id)) {
        onChange(selectedIds.filter((id) => id !== item.id));
        return;
      }
      onChange(mergeProjectContextSelection(selectedIds, [item.id]));
    },
    [disabled, onChange, selected, selectedIds],
  );

  const handleRemoveSelected = useCallback(
    (id: string) => {
      if (disabled) return;
      const next = selectedIds.filter((existing) => existing !== id);
      if (next.length === selectedIds.length) return;
      onChange(next);
    },
    [disabled, onChange, selectedIds],
  );

  if (!projectId) return null;

  const summaryBody = compact ? (
    <>
      <span
        className={[
          "project-context-picker__count-pill",
          selectedIds.length > 0 ? "project-context-picker__count-pill--active" : "",
        ]
          .filter(Boolean)
          .join(" ")}
      >
        {selectedCountLabel}
      </span>
      <span className="project-context-picker__summary-text">
        {compactSummaryText}
      </span>
    </>
  ) : (
    <>
      <strong>{selectedCountLabel}</strong>
      {contextQuery.isPending ? (
        <span>Loading project context...</span>
      ) : contextQuery.error ? (
        <span className="muted">Project context unavailable.</span>
      ) : items.length === 0 ? (
        <span>This project has no context nodes yet.</span>
      ) : selectedItems.length > 0 ? (
        <ul className="project-context-picker__chips">
          {selectedItems.map((item) => {
            const shortId = projectContextShortId(item.id);
            return (
              <li
                key={item.id}
                className="project-context-picker__chip"
                data-project-context-id={item.id}
              >
                <span className="project-context-picker__chip-title">
                  {item.title || "(untitled)"}
                </span>
                {shortId ? (
                  <span className="project-context-picker__chip-short-id muted">
                    · {shortId}
                  </span>
                ) : null}
                <button
                  type="button"
                  className="project-context-picker__chip-remove"
                  onClick={() => handleRemoveSelected(item.id)}
                  disabled={disabled}
                  aria-label={`Remove reference to ${item.title || "context node"}`}
                >
                  <svg
                    width="10"
                    height="10"
                    viewBox="0 0 10 10"
                    fill="none"
                    aria-hidden="true"
                  >
                    <path
                      d="M3 3l6 6M9 3l-6 6"
                      stroke="currentColor"
                      strokeWidth="1.4"
                      strokeLinecap="round"
                    />
                  </svg>
                </button>
              </li>
            );
          })}
        </ul>
      ) : (
        <span>Open the chooser to search the list.</span>
      )}
    </>
  );

  return (
    <section
      className={[
        "project-context-picker",
        compact ? "project-context-picker--compact" : "",
      ]
        .filter(Boolean)
        .join(" ")}
      aria-labelledby="task-context-picker-title"
    >
      <div className="project-context-picker__head">
        {compact ? (
          <span id="task-context-picker-title" className="project-context-picker__label">
            Project context
          </span>
        ) : (
          <div>
            <h3 id="task-context-picker-title">Context for this task</h3>
            <p>
              Reference project memory the agent may use. Add from the prompt
              with <kbd>#</kbd> or open the chooser. Backend resolves the full
              node memory at run time — chips here are display labels only.
            </p>
          </div>
        )}
        <button
          type="button"
          className="pc__btn-secondary project-context-picker__button"
          disabled={disabled}
          onClick={() => setChooserOpen(true)}
        >
          {compact ? "Choose" : "Choose context"}
        </button>
      </div>

      {compact ? (
        <p className="project-context-picker__lede">
          Type <kbd>#</kbd> in the prompt or choose nodes below.
        </p>
      ) : null}

      <div
        className="project-context-picker__summary"
        aria-live="polite"
        data-project-references-summary="true"
      >
        {summaryBody}
      </div>

      {chooserOpen ? (
        <Modal
          onClose={() => setChooserOpen(false)}
          labelledBy="task-context-chooser-title"
          describedBy="task-context-chooser-desc"
          size="wide"
        >
          <section className="panel modal-sheet modal-sheet--edit project-context-chooser pc">
            <div className="project-context-chooser__header">
              <div>
                <h2 id="task-context-chooser-title">Choose task context</h2>
                <p id="task-context-chooser-desc" className="muted">
                  Search project memory and select the nodes this task should
                  reference.
                </p>
              </div>
              <button
                type="button"
                className="pc__btn-ghost"
                onClick={() => setChooserOpen(false)}
              >
                Done
              </button>
            </div>

            <div className="pc__action-bar project-context-chooser__bar">
              <span className="pc__count">{selectedCountLabel}</span>
            </div>

            <div className="project-context-chooser__body">
              {contextQuery.isLoading ? (
                <div className="pc__skeleton" aria-hidden="true">
                  <div className="pd__shimmer pd__shimmer--card" />
                </div>
              ) : contextQuery.error ? (
                <div className="pd__inline-error" role="alert">
                  {contextQuery.error.message}
                </div>
              ) : items.length === 0 ? (
                <div className="pc__empty">
                  <p>No context nodes yet</p>
                  <span>
                    Import a memory file from the project context page first.
                  </span>
                </div>
              ) : (
                <ProjectContextListView
                  items={items}
                  selection={{
                    selectedIds: selected,
                    disabled,
                    onToggle: handleToggle,
                  }}
                />
              )}
            </div>

            <div className="project-context-chooser__footer">
              <button
                type="button"
                className="pc__btn-secondary"
                disabled={disabled || selectedIds.length === 0}
                onClick={() => onChange([])}
              >
                Clear selection
              </button>
              <button
                type="button"
                className="pc__btn-primary"
                onClick={() => setChooserOpen(false)}
              >
                Done
              </button>
            </div>
          </section>
        </Modal>
      ) : null}
    </section>
  );
}
