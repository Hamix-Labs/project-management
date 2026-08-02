import { formatLineRangeLabel } from "../repoFileRef";

/** Static export for blocksToHTMLLossy — no hooks / context. */
export function RepoFileEmbedExport(props: {
  path: string;
  lineStart: string;
  lineEnd: string;
}) {
  const start = props.lineStart ? parseInt(props.lineStart, 10) : undefined;
  const end = props.lineEnd ? parseInt(props.lineEnd, 10) : undefined;
  const range =
    Number.isFinite(start) && start
      ? formatLineRangeLabel(start, Number.isFinite(end) ? end : undefined)
      : null;

  return (
    <div
      className="file-embed"
      data-repo-file-embed="true"
      data-path={props.path}
      {...(props.lineStart ? { "data-line-start": props.lineStart } : {})}
      {...(props.lineEnd ? { "data-line-end": props.lineEnd } : {})}
    >
      <div className="file-embed-head">
        <span className="file-embed-path">{props.path}</span>
        {range ? <span className="file-embed-range">{range}</span> : null}
      </div>
    </div>
  );
}
