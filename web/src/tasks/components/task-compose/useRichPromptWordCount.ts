import { useCallback, useEffect, useState } from "react";
import type { Editor } from "@tiptap/core";

function countWords(text: string): number {
  const trimmed = text.trim();
  if (!trimmed) return 0;
  return trimmed.split(/\s+/).length;
}

/**
 * Tracks TipTap editor text length for the brief card word count.
 * Subscribes to editor transactions so the count stays live.
 */
export function useRichPromptWordCount() {
  const [wordCount, setWordCount] = useState(0);
  const [editor, setEditor] = useState<Editor | null>(null);

  const onEditorReady = useCallback((next: Editor | null) => {
    setEditor(next);
  }, []);

  useEffect(() => {
    if (!editor) {
      setWordCount(0);
      return;
    }
    const sync = () => setWordCount(countWords(editor.getText()));
    sync();
    editor.on("transaction", sync);
    return () => {
      editor.off("transaction", sync);
    };
  }, [editor]);

  return { wordCount, onEditorReady };
}
