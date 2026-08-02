import {
  formatLineRangeLabel,
  parseOptionalLine,
  sliceFileLines,
} from "../repoFileRef";
import { usePromptEditorRepo } from "../context/PromptEditorRepoContext";
import { useRepoFileContent } from "@/hooks/useRepoFileContent";

export type RepoFileEmbedViewProps = {
  path: string;
  lineStart: string;
  lineEnd: string;
};

function FileIcon() {
  return (
    <svg
      width="14"
      height="14"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
      aria-hidden="true"
    >
      <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z" />
      <path d="M14 2v6h6" />
    </svg>
  );
}

export function RepoFileEmbedView({
  path,
  lineStart: lineStartRaw,
  lineEnd: lineEndRaw,
}: RepoFileEmbedViewProps) {
  const { worktreeId } = usePromptEditorRepo();
  const lineStart = parseOptionalLine(lineStartRaw);
  const lineEnd = parseOptionalLine(lineEndRaw);
  const rangeLabel = formatLineRangeLabel(lineStart, lineEnd);
  const query = useRepoFileContent(path, worktreeId);

  const head = (
    <div className="file-embed-head">
      <FileIcon />
      <span className="file-embed-path">{path}</span>
      {rangeLabel ? (
        <span className="file-embed-range">{rangeLabel}</span>
      ) : null}
    </div>
  );

  if (!worktreeId?.trim()) {
    return (
      <div className="file-embed" data-repo-file-embed="true">
        {head}
        <div className="file-embed-body">File preview is unavailable.</div>
      </div>
    );
  }

  if (query.isLoading || query.isFetching) {
    return (
      <div className="file-embed" data-repo-file-embed="true" aria-busy="true">
        {head}
        <div className="file-embed-skeleton">
          <div className="file-embed-skeleton__line" />
          <div className="file-embed-skeleton__line" />
          <div className="file-embed-skeleton__line" style={{ width: "60%" }} />
        </div>
      </div>
    );
  }

  if (query.isError) {
    return (
      <div className="file-embed" data-repo-file-embed="true">
        {head}
        <div className="file-embed-body file-embed-body--error">
          {query.error instanceof Error ? query.error.message : "Load failed"}
          <div>
            <button
              type="button"
              className="file-embed-retry"
              onClick={() => void query.refetch()}
            >
              Retry
            </button>
          </div>
        </div>
      </div>
    );
  }

  const file = query.data;
  if (file === null || file === undefined) {
    return (
      <div className="file-embed" data-repo-file-embed="true">
        {head}
        <div className="file-embed-body">File preview is unavailable.</div>
      </div>
    );
  }

  if (file.binary) {
    return (
      <div className="file-embed" data-repo-file-embed="true">
        {head}
        <div className="file-embed-body">
          This file looks binary and cannot be previewed.
        </div>
      </div>
    );
  }

  const sliced = sliceFileLines(file.content, lineStart, lineEnd);

  return (
    <div className="file-embed" data-repo-file-embed="true">
      {head}
      <pre className="file-embed-code">
        {sliced.lines.map((line, i) => (
          <div key={sliced.start + i}>
            <span className="ln">{sliced.start + i}</span>
            {line}
          </div>
        ))}
      </pre>
      {file.truncated ? (
        <div className="file-embed-body">Preview truncated.</div>
      ) : null}
    </div>
  );
}
