import {
  BlockNoteSchema,
  defaultBlockSpecs,
  defaultInlineContentSpecs,
} from "@blocknote/core";
import { createReactInlineContentSpec } from "@blocknote/react";
import { createRepoFileEmbed } from "./blocks/repoFileEmbedSpec";
import { repoFileMentionLabel } from "./repoFileMentionLabel";

function parseOptionalInt(raw: string | null): number | undefined {
  if (raw == null || raw === "") return undefined;
  const n = parseInt(raw, 10);
  return Number.isFinite(n) ? n : undefined;
}

export const RepoFileMentionInline = createReactInlineContentSpec(
  {
    type: "repoFileMention" as const,
    propSchema: {
      path: { default: "" },
      lineStart: { default: "" },
      lineEnd: { default: "" },
    },
    content: "none",
  },
  {
    render: (props) => {
      const path = props.inlineContent.props.path;
      const lineStart = parseOptionalInt(props.inlineContent.props.lineStart);
      const lineEnd = parseOptionalInt(props.inlineContent.props.lineEnd);
      const label = repoFileMentionLabel({
        path,
        lineStart: lineStart ?? null,
        lineEnd: lineEnd ?? null,
      });
      return <span className="repo-file-chip">{label}</span>;
    },
    toExternalHTML: (props) => {
      const path = props.inlineContent.props.path;
      const lineStart = props.inlineContent.props.lineStart;
      const lineEnd = props.inlineContent.props.lineEnd;
      const start = parseOptionalInt(lineStart);
      const end = parseOptionalInt(lineEnd);
      const label = repoFileMentionLabel({
        path,
        lineStart: start ?? null,
        lineEnd: end ?? null,
      });
      const attrs: Record<string, string> = {
        "data-repo-file": "true",
        "data-path": path,
        class: "repo-file-chip",
      };
      if (start != null) attrs["data-line-start"] = String(start);
      if (end != null) attrs["data-line-end"] = String(end);
      return <span {...attrs}>{label}</span>;
    },
    parse: (element) => {
      if (
        element.tagName.toLowerCase() !== "span" ||
        element.getAttribute("data-repo-file") !== "true"
      ) {
        return undefined;
      }
      const path = element.getAttribute("data-path") ?? "";
      if (!path) return undefined;
      const lineStart = element.getAttribute("data-line-start") ?? "";
      const lineEnd = element.getAttribute("data-line-end") ?? "";
      return { path, lineStart, lineEnd };
    },
  },
);

export const promptEditorSchema = BlockNoteSchema.create({
  blockSpecs: {
    ...defaultBlockSpecs,
    repoFileEmbed: createRepoFileEmbed(),
  },
  inlineContentSpecs: {
    ...defaultInlineContentSpecs,
    repoFileMention: RepoFileMentionInline,
  },
});

export type PromptEditorSchema = typeof promptEditorSchema;
