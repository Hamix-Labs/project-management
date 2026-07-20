import { Link } from "react-router-dom";
import { useProjectDetailPrefetcher } from "@/app/hooks/usePrefetchOnIntent";
import type { Project } from "@/types";

/** Last path segment for repo labels (e.g. `C:/proj/hamix` → `hamix`). */
export function repositoryBasename(path: string): string {
  const normalized = path.trim().replace(/\\/g, "/").replace(/\/+$/, "");
  if (normalized === "") return path;
  const slash = normalized.lastIndexOf("/");
  return slash >= 0 ? normalized.slice(slash + 1) : normalized;
}

function formatDate(value: string) {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return "";
  return date.toLocaleDateString(undefined, { month: "short", day: "numeric" });
}

export function ProjectListRow({
  project,
  index,
  repositoryLabel,
}: {
  project: Project;
  index: number;
  repositoryLabel?: string;
}) {
  const isArchived = project.status === "archived";
  const displayName =
    project.is_default && repositoryLabel
      ? `${project.name} · ${repositoryLabel}`
      : project.name;
  const openLabel = `Open project ${displayName}`;
  const to = `/projects/${encodeURIComponent(project.id)}`;
  const prefetchProjectDetail = useProjectDetailPrefetcher();
  const onIntent = () => prefetchProjectDetail(project.id);

  return (
    <Link
      to={to}
      className={isArchived ? "pl__row pl__row--archived" : "pl__row pl__row--active"}
      style={{ animationDelay: `${index * 40}ms` }}
      aria-label={openLabel}
      onPointerEnter={onIntent}
      onFocus={onIntent}
    >
      <div className="pl__row-marker" aria-hidden="true" />
      <div className="pl__row-main">
        <span className="pl__row-name">{displayName}</span>
        <span className="pl__row-desc">
          {project.description || project.context_summary || "No description"}
        </span>
      </div>
      <div className="pl__row-meta">
        <span
          className={
            isArchived ? "pd__badge pd__badge--muted" : "pd__badge pd__badge--live"
          }
        >
          <span className="pd__badge-dot" aria-hidden="true" />
          {project.status}
        </span>
        <span className="pl__row-date">{formatDate(project.updated_at)}</span>
      </div>
      <svg className="pl__row-arrow" width="16" height="16" viewBox="0 0 16 16" fill="none" aria-hidden="true">
        <path d="M6 4l4 4-4 4" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round" />
      </svg>
    </Link>
  );
}

export function ProjectListSkeleton() {
  return (
    <div className="pl__list" aria-hidden="true">
      {Array.from({ length: 4 }).map((_, i) => (
        <div className="pl__row pl__row--skeleton" key={i}>
          <div className="pl__row-marker" />
          <div className="pl__row-main">
            <span className="pd__shimmer" style={{ width: `${60 - i * 8}%`, height: "0.9rem" }} />
            <span className="pd__shimmer" style={{ width: `${40 + i * 5}%`, height: "0.75rem" }} />
          </div>
          <div className="pl__row-meta">
            <span className="pd__shimmer" style={{ width: "3rem", height: "0.75rem" }} />
          </div>
        </div>
      ))}
    </div>
  );
}
