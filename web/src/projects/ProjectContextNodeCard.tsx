import type { ProjectContextItem } from "@/types";
import { ProjectContextItemEditor } from "./ProjectContextItemEditor";

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
  item: ProjectContextItem;
  index: number;
  existingTags: string[];
  saving: boolean;
  deleting: boolean;
  selected?: boolean;
  selectionDisabled?: boolean;
  onSave: (
    id: string,
    patch: {
      tag: string;
      title: string;
      description: string;
      body: string;
      pinned: boolean;
    },
  ) => void;
  onDelete: (id: string) => void;
  onToggleSelected?: (item: ProjectContextItem) => void;
};

export function ProjectContextNodeCard({
  item,
  index,
  existingTags,
  saving,
  deleting,
  selected = false,
  selectionDisabled = false,
  onSave,
  onDelete,
  onToggleSelected,
}: Props) {
  const hue = NODE_HUES[index % NODE_HUES.length];
  const nodeClass = onToggleSelected
    ? "pc__node pc__node--selectable"
    : "pc__node";
  return (
    <article
      className={nodeClass}
      style={
        { "--pc-hue": hue, animationDelay: `${index * 30}ms` } as React.CSSProperties
      }
    >
      <div className="pc__node-marker" aria-hidden="true" />
      {onToggleSelected ? (
        <label className="pc__node-select">
          <input
            type="checkbox"
            checked={selected}
            disabled={selectionDisabled}
            onChange={() => onToggleSelected(item)}
          />
          <span className="visually-hidden">Select {item.title}</span>
        </label>
      ) : null}
      <div className="pc__node-body">
        <h5 className="pc__node-title">{item.title}</h5>
        {item.description.trim() ? (
          <p className="pc__node-description">{item.description}</p>
        ) : null}
        <span className="pc__node-source">
          {item.created_by === "agent" ? "Agent" : "User"}
          {item.pinned ? " · Pinned" : ""}
        </span>
      </div>
      {!onToggleSelected ? (
        <div className="pc__node-actions">
          <ProjectContextItemEditor
            item={item}
            existingTags={existingTags}
            saving={saving}
            deleting={deleting}
            onSave={onSave}
            onDelete={onDelete}
          />
        </div>
      ) : null}
    </article>
  );
}
