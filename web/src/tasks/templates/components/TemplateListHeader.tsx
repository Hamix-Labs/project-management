type TemplateListHeaderProps = {
  allSelected: boolean;
  someSelected: boolean;
  onToggleSelectAll: () => void;
};

export function TemplateListHeader({
  allSelected,
  someSelected,
  onToggleSelectAll,
}: TemplateListHeaderProps) {
  return (
    <div className="templates-list-header" role="row">
      <div className="templates-list-header__select">
        <input
          type="checkbox"
          className="templates-list-header__checkbox"
          checked={allSelected}
          ref={(el) => {
            if (el) el.indeterminate = someSelected;
          }}
          onChange={onToggleSelectAll}
          aria-label={allSelected ? "Deselect all templates" : "Select all templates"}
          data-testid="template-list-select-all"
        />
      </div>
      <span className="templates-list-header__label" role="columnheader">
        Title
      </span>
      <span
        className="templates-list-header__label templates-list-header__label--instances"
        role="columnheader"
      >
        Instances
      </span>
      <span
        className="templates-list-header__label templates-list-header__label--actions"
        role="columnheader"
      >
        Actions
      </span>
    </div>
  );
}
