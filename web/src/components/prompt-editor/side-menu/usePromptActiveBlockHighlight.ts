import { useCallback, useEffect, useRef } from "react";
import { useEditorChange } from "@blocknote/react";
import {
  findPromptBlockElement,
  PROMPT_BLOCK_ACTIVE_ATTR,
} from "./promptBlockElement";

function clearActiveAttr(editorDom: HTMLElement | null | undefined, blockId: string) {
  findPromptBlockElement(editorDom, blockId)?.removeAttribute(
    PROMPT_BLOCK_ACTIVE_ATTR,
  );
}

function setActiveAttr(editorDom: HTMLElement | null | undefined, blockId: string) {
  findPromptBlockElement(editorDom, blockId)?.setAttribute(
    PROMPT_BLOCK_ACTIVE_ATTR,
    "true",
  );
}

/**
 * Stamps {@link PROMPT_BLOCK_ACTIVE_ATTR} on the live block containers for the
 * given ids, and clears it from any previously highlighted blocks.
 *
 * Re-applies on every editor change while ids are active: ProseMirror replaces
 * block DOM nodes on move/edit, which would otherwise drop the attribute.
 */
export function usePromptActiveBlockHighlight(
  editorDom: HTMLElement | null,
  activeBlockIds: readonly string[],
) {
  const prevIdsRef = useRef<string[]>([]);
  const activeKey = activeBlockIds.join("\0");

  const apply = useCallback(() => {
    const prev = prevIdsRef.current;
    const next = activeKey.length === 0 ? [] : activeKey.split("\0");

    for (const id of prev) {
      if (!next.includes(id)) {
        clearActiveAttr(editorDom, id);
      }
    }
    for (const id of next) {
      setActiveAttr(editorDom, id);
    }

    prevIdsRef.current = next;
  }, [activeKey, editorDom]);

  useEffect(() => {
    apply();
    return () => {
      for (const id of prevIdsRef.current) {
        clearActiveAttr(editorDom, id);
      }
      prevIdsRef.current = [];
    };
  }, [apply, editorDom]);

  // ProseMirror may swap the node under us without React noticing.
  useEditorChange(apply);
}
