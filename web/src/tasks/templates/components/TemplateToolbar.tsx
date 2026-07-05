import { CustomSelect, type CustomSelectOption } from "@/components/custom-select";

export type TemplateSortKey = "recent" | "title" | "runs";

export const TEMPLATE_SORT_LABELS: Record<TemplateSortKey, string> = {
  recent: "Recently updated",
  title: "Title (A–Z)",
  runs: "Most used",
};

const TEMPLATE_SORT_OPTIONS: CustomSelectOption[] = (
  Object.keys(TEMPLATE_SORT_LABELS) as TemplateSortKey[]
).map((key) => ({
  value: key,
  label: TEMPLATE_SORT_LABELS[key],
}));

type TemplateToolbarProps = {
  searchInput: string;
  sort: TemplateSortKey;
  onSearchChange: (value: string) => void;
  onSortChange: (sort: TemplateSortKey) => void;
};

function SearchIcon() {
  return (
    <svg
      className="templates-toolbar__search-icon"
      width="16"
      height="16"
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
  );
}

function SlidersIcon() {
  return (
    <svg width="14" height="14" viewBox="0 0 24 24" fill="none" aria-hidden="true">
      <path d="M21 4H14" stroke="currentColor" strokeWidth="2" strokeLinecap="round" />
      <path d="M10 4H3" stroke="currentColor" strokeWidth="2" strokeLinecap="round" />
      <path d="M21 12H12" stroke="currentColor" strokeWidth="2" strokeLinecap="round" />
      <path d="M8 12H3" stroke="currentColor" strokeWidth="2" strokeLinecap="round" />
      <path d="M21 20H16" stroke="currentColor" strokeWidth="2" strokeLinecap="round" />
      <path d="M12 20H3" stroke="currentColor" strokeWidth="2" strokeLinecap="round" />
      <path d="M14 2v4" stroke="currentColor" strokeWidth="2" strokeLinecap="round" />
      <path d="M8 10v4" stroke="currentColor" strokeWidth="2" strokeLinecap="round" />
      <path d="M16 18v4" stroke="currentColor" strokeWidth="2" strokeLinecap="round" />
    </svg>
  );
}

export function TemplateToolbar({
  searchInput,
  sort,
  onSearchChange,
  onSortChange,
}: TemplateToolbarProps) {
  return (
    <div className="templates-toolbar" role="search" aria-label="Filter templates">
      <div className="templates-toolbar__search-wrap">
        <SearchIcon />
        <label htmlFor="templates-search" className="visually-hidden">
          Search templates
        </label>
        <input
          id="templates-search"
          type="search"
          className="templates-toolbar__search"
          placeholder="Search templates…"
          autoComplete="off"
          value={searchInput}
          onChange={(e) => onSearchChange(e.target.value)}
        />
        {searchInput ? (
          <button
            type="button"
            className="templates-toolbar__search-clear"
            aria-label="Clear search"
            onClick={() => onSearchChange("")}
          >
            ×
          </button>
        ) : null}
      </div>
      <div className="templates-toolbar__sort-field">
        <CustomSelect
          id="templates-sort"
          label="Sort templates"
          listboxName="Sort templates"
          compact
          dropdownVariant="toolbar"
          dropdownMinWidth={220}
          value={sort}
          options={TEMPLATE_SORT_OPTIONS}
          onChange={(value) => onSortChange(value as TemplateSortKey)}
          leadingIcon={<SlidersIcon />}
        />
      </div>
    </div>
  );
}
