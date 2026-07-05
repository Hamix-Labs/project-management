type TemplateEmptyStateProps = {
  hasFilters: boolean;
  onClearFilters: () => void;
  onNewTemplate: () => void;
};

function SearchIcon() {
  return (
    <svg width="20" height="20" viewBox="0 0 14 14" fill="none" aria-hidden="true">
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

export function TemplateEmptyState({
  hasFilters,
  onClearFilters,
  onNewTemplate,
}: TemplateEmptyStateProps) {
  return (
    <div className="templates-empty-state">
      <div className="templates-empty-state__icon">
        <SearchIcon />
      </div>
      <div className="templates-empty-state__text">
        <p className="templates-empty-state__title">
          {hasFilters ? "No templates found" : "No templates yet"}
        </p>
        <p className="templates-empty-state__desc">
          {hasFilters
            ? "Try adjusting your search or filters."
            : "Create your first task template to get started."}
        </p>
      </div>
      {hasFilters ? (
        <button type="button" className="secondary" onClick={onClearFilters}>
          Clear filters
        </button>
      ) : (
        <button type="button" className="templates-page-header__new-btn" onClick={onNewTemplate}>
          New template
        </button>
      )}
    </div>
  );
}
