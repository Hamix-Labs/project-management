import { createReactBlockSpec } from "@blocknote/react";
import { RepoFileEmbedExport } from "./RepoFileEmbedExport";
import { RepoFileEmbedView } from "./RepoFileEmbedView";

export const createRepoFileEmbed = createReactBlockSpec(
  {
    type: "repoFileEmbed" as const,
    propSchema: {
      path: { default: "" },
      lineStart: { default: "" },
      lineEnd: { default: "" },
    },
    content: "none",
  },
  {
    render: (props) => (
      <RepoFileEmbedView
        path={props.block.props.path}
        lineStart={props.block.props.lineStart}
        lineEnd={props.block.props.lineEnd}
      />
    ),
    toExternalHTML: (props) => (
      <RepoFileEmbedExport
        path={props.block.props.path}
        lineStart={props.block.props.lineStart}
        lineEnd={props.block.props.lineEnd}
      />
    ),
    parse: (element) => {
      if (
        element.tagName.toLowerCase() !== "div" ||
        element.getAttribute("data-repo-file-embed") !== "true"
      ) {
        return undefined;
      }
      const path = element.getAttribute("data-path") ?? "";
      if (!path) return undefined;
      return {
        path,
        lineStart: element.getAttribute("data-line-start") ?? "",
        lineEnd: element.getAttribute("data-line-end") ?? "",
      };
    },
  },
);
