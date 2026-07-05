import type { ReactNode } from "react";
import { TEMPLATE_CATEGORY_LABELS } from "../templateCategories";

type TemplateTagFiltersProps = {
  activeTag: string;
  dynamicTags: string[];
  onTagChange: (tag: string) => void;
};

function FilterChip({
  active,
  onClick,
  children,
}: {
  active: boolean;
  onClick: () => void;
  children: ReactNode;
}) {
  return (
    <button
      type="button"
      className={[
        "templates-tag-chip",
        active ? "templates-tag-chip--active" : "",
      ]
        .filter(Boolean)
        .join(" ")}
      aria-pressed={active}
      onClick={onClick}
    >
      {children}
    </button>
  );
}

export function TemplateTagFilters({
  activeTag,
  dynamicTags,
  onTagChange,
}: TemplateTagFiltersProps) {
  return (
    <div className="templates-tag-filters" role="group" aria-label="Filter by tag">
      <FilterChip active={activeTag === "all"} onClick={() => onTagChange("all")}>
        All
      </FilterChip>
      {TEMPLATE_CATEGORY_LABELS.map((label) => (
        <FilterChip
          key={label}
          active={activeTag === label}
          onClick={() => onTagChange(label)}
        >
          {label}
        </FilterChip>
      ))}
      {dynamicTags.map((tag) => (
        <FilterChip key={tag} active={activeTag === tag} onClick={() => onTagChange(tag)}>
          {tag}
        </FilterChip>
      ))}
    </div>
  );
}
