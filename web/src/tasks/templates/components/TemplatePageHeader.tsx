type TemplatePageHeaderProps = {
  onNewTemplate: () => void;
};

function PlusIcon() {
  return (
    <svg width="16" height="16" viewBox="0 0 16 16" fill="none" aria-hidden="true">
      <path
        d="M8 3v10M3 8h10"
        stroke="currentColor"
        strokeWidth="1.75"
        strokeLinecap="round"
      />
    </svg>
  );
}

export function TemplatePageHeader({ onNewTemplate }: TemplatePageHeaderProps) {
  return (
    <header className="templates-page-header">
      <div className="templates-page-header__text">
        <h1 id="task-templates-heading" className="templates-page-header__title">
          Task templates
        </h1>
        <p className="templates-page-header__subtitle">
          Reusable task definitions. Select a few and spin up tasks in bulk.
        </p>
      </div>
      <button
        type="button"
        className="templates-page-header__new-btn"
        onClick={onNewTemplate}
      >
        <PlusIcon />
        New template
      </button>
    </header>
  );
}
