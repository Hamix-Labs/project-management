type LoadMoreRowsProps = {
  shown: number;
  total: number;
  itemLabel: string;
  onLoadMore: () => void;
};

export function LoadMoreRows({
  shown,
  total,
  itemLabel,
  onLoadMore,
}: LoadMoreRowsProps) {
  if (shown >= total) {
    return (
      <p className="task-attempt-count muted">
        All {total} {itemLabel} shown.
      </p>
    );
  }
  return (
    <div className="task-attempt-load-more">
      <p className="task-attempt-count muted">
        {shown} of {total} {itemLabel}
      </p>
      <button type="button" className="secondary" onClick={onLoadMore}>
        Load more
      </button>
    </div>
  );
}
