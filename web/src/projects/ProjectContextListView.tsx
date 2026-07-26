import { useMemo, useState } from "react";
import type { ProjectContextItem } from "@/types";
import { ProjectContextNodeCard } from "./ProjectContextNodeCard";
import { groupProjectContextByTag } from "./projectContextTags";

type Props = {
  items: ProjectContextItem[];
  nodeSaving?: boolean;
  nodeDeleting?: boolean;
  onSaveNode?: (
    id: string,
    patch: {
      tag: string;
      title: string;
      description: string;
      body: string;
      pinned: boolean;
    },
  ) => void;
  onDeleteNode?: (id: string) => void;
  selection?: {
    selectedIds: Set<string>;
    disabled?: boolean;
    onToggle: (item: ProjectContextItem) => void;
  };
  onImportMemory?: () => void;
  showImportAction?: boolean;
};

export function ProjectContextListView({
  items,
  nodeSaving,
  nodeDeleting,
  onSaveNode,
  onDeleteNode,
  selection,
  onImportMemory,
  showImportAction = false,
}: Props) {
  const [nodeQuery, setNodeQuery] = useState("");
  const filteredItems = useMemo(() => {
    const query = nodeQuery.trim().toLowerCase();
    if (!query) return items;
    return items.filter((item) =>
      [item.title, item.description, item.body, item.tag]
        .join(" ")
        .toLowerCase()
        .includes(query),
    );
  }, [items, nodeQuery]);
  const groups = useMemo(
    () => groupProjectContextByTag(filteredItems),
    [filteredItems],
  );
  const resultLabel =
    nodeQuery.trim().length > 0
      ? `${filteredItems.length} of ${items.length}`
      : `${items.length}`;

  return (
    <div className="pc__list-view">
      <div className="pc__list-bar">
        {showImportAction && onImportMemory ? (
          <button
            type="button"
            className="pc__btn-primary"
            onClick={onImportMemory}
          >
            <span aria-hidden="true">+</span>
            Import memory file
          </button>
        ) : null}
        <label className="pc__search">
          <span className="visually-hidden">Search memory nodes</span>
          <svg
            className="pc__search-icon"
            width="14"
            height="14"
            viewBox="0 0 14 14"
            fill="none"
            aria-hidden="true"
          >
            <circle
              cx="6"
              cy="6"
              r="4.5"
              stroke="currentColor"
              strokeWidth="1.2"
            />
            <path
              d="M9.5 9.5L13 13"
              stroke="currentColor"
              strokeWidth="1.2"
              strokeLinecap="round"
            />
          </svg>
          <input
            value={nodeQuery}
            onChange={(event) => setNodeQuery(event.target.value)}
            placeholder="Search by title, description, body, or tag..."
          />
          <span className="pc__count">{resultLabel}</span>
        </label>
      </div>

      {filteredItems.length === 0 ? (
        <div className="pc__empty">
          <p>No matching nodes</p>
          <span>Try a different search term or clear the filter.</span>
        </div>
      ) : (
        <div className="pc__tag-groups">
          {groups.map((group) => (
            <section key={group.label} className="pc__tag-group">
              <header className="pc__tag-group-header">
                <h3 className="pc__tag-group-title">{group.label}</h3>
                <span className="pc__tag-group-count">{group.items.length}</span>
              </header>
              <ul className="pc__tag-group-list">
                {group.items.map((item, i) => (
                  <li key={item.id}>
                    <ProjectContextNodeCard
                      item={item}
                      index={i}
                      existingTags={collectTagsExcept(items, item.id)}
                      saving={nodeSaving ?? false}
                      deleting={nodeDeleting ?? false}
                      selected={selection?.selectedIds.has(item.id) ?? false}
                      selectionDisabled={selection?.disabled}
                      onSave={onSaveNode ?? (() => undefined)}
                      onDelete={onDeleteNode ?? (() => undefined)}
                      onToggleSelected={selection?.onToggle}
                    />
                  </li>
                ))}
              </ul>
            </section>
          ))}
        </div>
      )}
    </div>
  );
}

function collectTagsExcept(
  items: ProjectContextItem[],
  excludeId: string,
): string[] {
  const seen = new Map<string, string>();
  for (const item of items) {
    if (item.id === excludeId) continue;
    const display = item.tag.trim();
    const key = display.toLowerCase();
    if (!key || seen.has(key)) continue;
    seen.set(key, display);
  }
  return [...seen.values()];
}
