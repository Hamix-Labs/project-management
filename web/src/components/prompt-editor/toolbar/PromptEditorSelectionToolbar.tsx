import {
  BasicTextStyleButton,
  BlockTypeSelect,
  ColorStyleButton,
  CreateLinkButton,
  FileCaptionButton,
  FileDeleteButton,
  FileDownloadButton,
  FilePreviewButton,
  FileRenameButton,
  FileReplaceButton,
  FormattingToolbar,
  TableCellMergeButton,
} from "@blocknote/react";

/**
 * Selection toolbar for the Prompt IDE, stacked as vertical sections rather
 * than BlockNote's single horizontal row.
 *
 * Rows are plain `div`s so the panel can lay out as a grid, which is safe for
 * the underlying Ariakit toolbar: it collects its items through React context,
 * not by walking direct DOM children. Each row collapses via `:empty` in
 * `app-prompt-editor-blocknote-menus.css` because every BlockNote button
 * renders `null` when it does not apply to the current selection.
 *
 * Alignment, indent/outdent, and comment buttons are deliberately absent —
 * block placement belongs to the block menu, and commenting is not a Prompt
 * IDE feature.
 */
export function PromptEditorSelectionToolbar() {
  return (
    <FormattingToolbar>
      <div className="prompt-selection-toolbar__row prompt-selection-toolbar__row--block-type">
        <BlockTypeSelect />
      </div>
      <div className="prompt-selection-toolbar__row prompt-selection-toolbar__row--format">
        <ColorStyleButton />
        <BasicTextStyleButton basicTextStyle="bold" />
        <BasicTextStyleButton basicTextStyle="italic" />
        <BasicTextStyleButton basicTextStyle="underline" />
        <CreateLinkButton />
        <BasicTextStyleButton basicTextStyle="strike" />
        <BasicTextStyleButton basicTextStyle="code" />
      </div>
      <div className="prompt-selection-toolbar__row prompt-selection-toolbar__row--block-actions">
        <TableCellMergeButton />
        <FileCaptionButton />
        <FileReplaceButton />
        <FileRenameButton />
        <FileDeleteButton />
        <FileDownloadButton />
        <FilePreviewButton />
      </div>
    </FormattingToolbar>
  );
}
