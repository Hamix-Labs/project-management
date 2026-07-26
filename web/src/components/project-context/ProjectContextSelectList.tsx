import { useMemo, useState, type CSSProperties } from "react";
import type { ProjectContextItem } from "@/types";
import { projectContextShortId } from "@/lib/projectContextRefs";

const NODE_HUES = [
  "248, 63%",
  "160, 60%",
  "330, 55%",
  "38, 75%",
  "200, 65%",
  "280, 50%",
  "15, 65%",
  "175, 55%",
];

type Props = {
  items: ProjectContextItem[];
  selectedIds: Set<string>;
  disabled?: boolean;
  onToggle: (item: ProjectContextItem) => void;
};

/**
 * Selection-only context list for task compose choosers. Lives in the
 * inner ring so `components/` does not import the projects vertical
 * (full edit/delete cards stay under `projects/ProjectContextListView`).
 */
export function ProjectContextSelectList({
  items,
  selectedIds,
  disabled,
  onToggle,
}: Props) {
  const [nodeQuery, setNodeQuery] = useState("");
  const filteredItems = useMemo(() => {
    const query = nodeQuery.trim().toLowerCase();
    if (!query) return items;
    return items.filter((item) =>
      [item.title, item.body, item.tag]
        .join(" ")
        .toLowerCase()
        .includes(query),
    );
  }, [items, nodeQuery]);
  const resultLabel =
    nodeQuery.trim().length > 0
      ? `${filteredItems.length} of ${items.length}`
      : `${items.length}`;

  return (
    <div className="pc__list-view">
      <div className="pc__list-bar">
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
            <circle cx="6" cy="6" r="4.5" stroke="currentColor" strokeWidth="1.2" />
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
            placeholder="Search by title, body, or tag..."
          />
        </label>
        <span className="pc__count">{resultLabel}</span>
      </div>

      {filteredItems.length === 0 ? (
        <div className="pc__empty">
          <p>No matching nodes</p>
          <span>Try a different search term or clear the filter.</span>
        </div>
      ) : (
        <div className="pc__node-grid">
          {filteredItems.map((item, i) => {
            const hue = NODE_HUES[i % NODE_HUES.length];
            const shortId = projectContextShortId(item.id);
            return (
              <article
                key={item.id}
                className="pc__node pc__node--selectable"
                style={
                  {
                    "--pc-hue": hue,
                    animationDelay: `${i * 30}ms`,
                  } as CSSProperties
                }
              >
                <div className="pc__node-marker" aria-hidden="true" />
                <label className="pc__node-select">
                  <input
                    type="checkbox"
                    checked={selectedIds.has(item.id)}
                    disabled={disabled}
                    onChange={() => onToggle(item)}
                  />
                  <span className="visually-hidden">
                    Select {item.title || "context node"}
                  </span>
                </label>
                <div className="pc__node-body">
                  <h5 className="pc__node-title">{item.title || "(untitled)"}</h5>
                  <span className="pc__node-source">
                    {item.created_by === "agent" ? "Agent" : "User"}
                    {item.pinned ? " · Pinned" : ""}
                    {shortId ? ` · ${shortId}` : ""}
                  </span>
                </div>
                <span className="pc__node-kind">{item.tag}</span>
              </article>
            );
          })}
        </div>
      )}
    </div>
  );
}
