import { useQuery } from "@tanstack/react-query";
import { Link } from "react-router-dom";
import { listProjectContext } from "@/api";
import { projectQueryKeys } from "./queryKeys";

const ENTRY_META_LIMIT = 100;

type Props = {
  projectId: string;
};

export function ProjectContextEntryCard({ projectId }: Props) {
  const contextQuery = useQuery({
    queryKey: projectQueryKeys.contextEntryMeta(projectId),
    queryFn: ({ signal }) =>
      listProjectContext(projectId, { signal, limit: ENTRY_META_LIMIT }),
    enabled: Boolean(projectId),
  });
  const n = contextQuery.data?.items?.length ?? 0;
  const atCap = n >= ENTRY_META_LIMIT;
  const label =
    contextQuery.isLoading || contextQuery.isFetching
      ? "Loading nodes…"
      : atCap
        ? `${ENTRY_META_LIMIT}+ nodes`
        : `${n} ${n === 1 ? "node" : "nodes"}`;

  return (
    <Link
      to={`/projects/${encodeURIComponent(projectId)}/context`}
      className="pd__context-row"
      aria-labelledby="pd-context-title"
      aria-label={`Open project context. ${label}`}
    >
      <span className="pd__context-icon" aria-hidden="true">
        <svg width="16" height="16" viewBox="0 0 24 24" fill="none">
          <ellipse
            cx="12"
            cy="5"
            rx="8"
            ry="3"
            stroke="currentColor"
            strokeWidth="1.75"
          />
          <path
            d="M4 5v6c0 1.66 3.58 3 8 3s8-1.34 8-3V5"
            stroke="currentColor"
            strokeWidth="1.75"
            strokeLinecap="round"
          />
          <path
            d="M4 11v6c0 1.66 3.58 3 8 3s8-1.34 8-3v-6"
            stroke="currentColor"
            strokeWidth="1.75"
            strokeLinecap="round"
          />
        </svg>
      </span>
      <span className="pd__context-text">
        <span id="pd-context-title" className="pd__context-title">
          Project context
        </span>
        <span className="pd__context-desc">
          Memory nodes, decisions, and constraints
        </span>
      </span>
      <span className="pd__context-meta" aria-live="polite">
        {label}
      </span>
      <svg
        className="pd__context-arrow"
        width="14"
        height="14"
        viewBox="0 0 16 16"
        fill="none"
        aria-hidden="true"
      >
        <path
          d="M6 4l4 4-4 4"
          stroke="currentColor"
          strokeWidth="1.5"
          strokeLinecap="round"
          strokeLinejoin="round"
        />
      </svg>
    </Link>
  );
}
