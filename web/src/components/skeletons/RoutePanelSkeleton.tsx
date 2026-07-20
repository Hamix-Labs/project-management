/** Shell-compatible placeholder while a lazy route chunk downloads. */
export function RoutePanelSkeleton() {
  return (
    <section className="panel route-panel-skeleton" aria-busy="true">
      <div
        className="stack"
        role="status"
        aria-label="Loading page"
      >
        <span className="skeleton-block skeleton-block--detail-title" />
        <span className="skeleton-block skeleton-block--detail-line" />
        <span className="skeleton-block skeleton-block--detail-line-short" />
        <span className="skeleton-block skeleton-block--detail-line" />
      </div>
    </section>
  );
}
