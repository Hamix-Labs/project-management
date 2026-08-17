import type { EditorView } from "@tiptap/pm/view";

function hasImageFile(files: FileList | undefined | null): boolean {
  return Boolean(files && Array.from(files).some((f) => f.type.startsWith("image/")));
}

function hasImageItem(items: DataTransferItemList | undefined | null): boolean {
  return Boolean(
    items && Array.from(items).some((item) => item.type.startsWith("image/")),
  );
}

/**
 * Prompts store agent HTML, not raster attachments. Ignore image paste/drop
 * so the schema never grows an upload node.
 */
export function handlePromptImagePaste(
  _view: EditorView,
  event: ClipboardEvent,
): boolean {
  if (hasImageFile(event.clipboardData?.files) || hasImageItem(event.clipboardData?.items)) {
    event.preventDefault();
    return true;
  }
  return false;
}

export function handlePromptImageDrop(_view: EditorView, event: DragEvent): boolean {
  if (hasImageFile(event.dataTransfer?.files)) {
    event.preventDefault();
    return true;
  }
  return false;
}
