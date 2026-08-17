import type { Editor } from "@tiptap/react";
import { EditorContent, EditorContext } from "@tiptap/react";
import { DragContextMenu } from "@/components/tiptap-ui/drag-context-menu";
import { EmojiDropdownMenu } from "@/components/tiptap-ui/emoji-dropdown-menu";
import { SlashDropdownMenu } from "@/components/tiptap-ui/slash-dropdown-menu";
import { NotionToolbarFloating } from "@/components/tiptap-templates/notion-like/notion-like-editor-toolbar-floating";
import { TableHandle } from "@/components/tiptap-node/table-node/ui/table-handle";
import {
  TableSelectionOverlay,
  type TableSelectionOverlayProps,
} from "@/components/tiptap-node/table-node/ui/table-selection-overlay";
import { TableCellHandleMenu } from "@/components/tiptap-node/table-node/ui/table-cell-handle-menu";
import { TableExtendRowColumnButtons } from "@/components/tiptap-node/table-node/ui/table-extend-row-column-button";
import { useUiEditorState } from "@/hooks/use-ui-editor-state";
import { buildHamixSlashMenuConfig } from "./hamixSlashMenu";
import { useMemo, type ReactNode } from "react";
import "./hamixNotionNodeStyles";

const HamixTableCellMenu: NonNullable<TableSelectionOverlayProps["cellMenu"]> = (
  props,
) => (
  <TableCellHandleMenu
    editor={props.editor}
    onMouseDown={(e) => props.onResizeStart?.("br")(e)}
  />
);

type Props = {
  editor: Editor | null;
  toolbar?: "full" | "none";
  menuRight?: ReactNode;
  onAiTrigger?: (msg: string) => void;
};

/** Notion-like overlays around one Hamix `useEditor` instance. No Cloud collab/AI. */
export function HamixNotionEditorSurface({
  editor,
  toolbar = "full",
  menuRight,
  onAiTrigger,
}: Props) {
  const { isDragging } = useUiEditorState(editor);
  const slashConfig = useMemo(
    () => buildHamixSlashMenuConfig(onAiTrigger),
    [onAiTrigger],
  );

  return (
    <EditorContext.Provider value={{ editor }}>
      {menuRight ? <div className="rich-prompt-meta">{menuRight}</div> : null}
      <EditorContent
        editor={editor}
        role="presentation"
        className="notion-like-editor-content"
        style={{ cursor: isDragging ? "grabbing" : "auto" }}
      >
        {editor ? <DragContextMenu /> : null}
        {editor ? <EmojiDropdownMenu /> : null}
        {editor ? <SlashDropdownMenu config={slashConfig} /> : null}
        {toolbar === "full" && editor ? <NotionToolbarFloating /> : null}
      </EditorContent>
      {editor ? (
        <>
          <TableExtendRowColumnButtons />
          <TableHandle />
          <TableSelectionOverlay
            showResizeHandles={true}
            cellMenu={HamixTableCellMenu}
          />
        </>
      ) : null}
    </EditorContext.Provider>
  );
}
