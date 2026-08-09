import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { SideMenuExtension } from "@blocknote/core/extensions";
import {
  DragHandleButton,
  GenericPopover,
  SideMenu,
  useBlockNoteEditor,
  useEditorChange,
  useEditorDOMElement,
  useExtensionState,
  type GenericPopoverReference,
} from "@blocknote/react";
import { decideActiveBlockIds } from "./decideActiveBlockIds";
import { PromptEditorAddBlockButton } from "./PromptEditorAddBlockButton";
import { PromptEditorDragHandleMenu } from "./PromptEditorDragHandleMenu";
import { PromptDragHandleMenuOpenProvider } from "./promptDragHandleMenuOpenContext";
import { promptSideMenuAnchorRect } from "./promptSideMenuAnchor";
import { usePromptActiveBlockHighlight } from "./usePromptActiveBlockHighlight";
import { usePromptBlockDragState } from "./usePromptBlockDragState";

/**
 * Prompt IDE replacement for BlockNote's `SideMenuController`.
 *
 * The stock controller anchors to the DOM node it resolved when the hovered
 * block id last changed. A block keeps its id when it moves, so nothing
 * recomputes that node and floating UI keeps positioning against the detached
 * original — the add/drag buttons stay where the block used to be. Here the
 * anchor is a virtual reference that re-reads the block's rect on every
 * position pass, so correctness never depends on a re-render or a remount.
 *
 * Visibility comes from the extension's own `show` state plus a drag flag; the
 * menu stays mounted while a drag is in flight so the drag handle survives as
 * the drag source and still emits `dragend`.
 *
 * The buttons are passed explicitly so the add-block affordance can be replaced.
 * The drag handle stays BlockNote's; its menu is prompt-owned so we can observe
 * menu-open (for the active-block highlight) and so #156 can extend the items.
 */
export function PromptEditorSideMenu({
  editorHost,
}: {
  editorHost: HTMLElement | null;
}) {
  const editor = useBlockNoteEditor();
  const editorDom = useEditorDOMElement();
  const dragging = usePromptBlockDragState(editorHost);
  const { show, blockId } = useExtensionState(SideMenuExtension, {
    selector: (state) => ({
      show: state?.show ?? false,
      blockId: state?.block?.id as string | undefined,
    }),
  });

  const open = show && blockId !== undefined;
  const [dragHandleMenuOpen, setDragHandleMenuOpen] = useState(false);

  // Side-menu dismiss (hover leave, Escape via BlockNote, delete) must clear
  // the highlight even if the Ariakit beacon teardown is delayed a frame.
  useEffect(() => {
    if (!open) {
      setDragHandleMenuOpen(false);
    }
  }, [open]);

  const [selectionBlockIds, setSelectionBlockIds] = useState<
    string[] | undefined
  >(undefined);

  const readSelectionBlockIds = useCallback(() => {
    if (!dragHandleMenuOpen && !dragging) {
      setSelectionBlockIds((prev) => (prev === undefined ? prev : undefined));
      return;
    }
    const ids = editor.getSelection()?.blocks.map((block) => block.id);
    setSelectionBlockIds((prev) => {
      if (
        prev !== undefined &&
        ids !== undefined &&
        prev.length === ids.length &&
        prev.every((id, index) => id === ids[index])
      ) {
        return prev;
      }
      return ids;
    });
  }, [dragHandleMenuOpen, dragging, editor]);

  useEffect(() => {
    readSelectionBlockIds();
  }, [readSelectionBlockIds]);

  useEditorChange(readSelectionBlockIds);

  const activeBlockIds = useMemo(
    () =>
      decideActiveBlockIds({
        menuOpen: dragHandleMenuOpen,
        dragging,
        targetBlockId: blockId,
        selectionBlockIds,
      }),
    [blockId, dragHandleMenuOpen, dragging, selectionBlockIds],
  );

  usePromptActiveBlockHighlight(editorDom ?? null, activeBlockIds);

  const updateRef = useRef<(() => void) | null>(null);
  const whileElementsMounted = useCallback(
    (_reference: unknown, _floating: unknown, update: () => void) => {
      updateRef.current = update;
      return () => {
        updateRef.current = null;
      };
    },
    [],
  );

  // A block can move without the hovered block id changing (drag & drop, undo,
  // a delete above it), which leaves the menu anchored to stale coordinates.
  // Repositioning is enough — remounting would restart the fade and tear down
  // the drag handle mid-drag. Deferring a frame both coalesces bursts and keeps
  // the reposition out of `dragstart`, where moving the drag handle's ancestor
  // would cancel the drag.
  const repositionFrame = useRef(0);
  const scheduleReposition = useCallback(() => {
    if (repositionFrame.current) {
      return;
    }
    repositionFrame.current = requestAnimationFrame(() => {
      repositionFrame.current = 0;
      updateRef.current?.();
    });
  }, []);

  useEditorChange(scheduleReposition);

  useEffect(
    () => () => {
      if (repositionFrame.current) {
        cancelAnimationFrame(repositionFrame.current);
      }
    },
    [],
  );

  // A finished drag moved the block without changing the hovered block id, so
  // nothing else would trigger a position pass.
  useEffect(() => {
    if (!dragging) {
      scheduleReposition();
    }
  }, [dragging, scheduleReposition]);

  const reference = useMemo<GenericPopoverReference | undefined>(() => {
    if (!editorDom || blockId === undefined) {
      return undefined;
    }

    const getBoundingClientRect = () =>
      promptSideMenuAnchorRect(editorDom, blockId) ?? new DOMRect();
    const gutter = editorDom.firstElementChild;

    if (!gutter) {
      return { element: undefined, getBoundingClientRect };
    }

    return {
      // The block group is the stable scroll-parent descendant FloatingUI needs
      // for auto-updates; the rect itself always comes from the block.
      element: gutter,
      cacheMountedBoundingClientRect: false,
      getBoundingClientRect,
    };
  }, [editorDom, blockId]);

  const floatingUIOptions = useMemo(
    () => ({
      useFloatingOptions: {
        open,
        placement: "left-start" as const,
        whileElementsMounted,
      },
      // The stylesheet owns the fade so it can also cover the dragging state.
      useTransitionStylesProps: { duration: 0 },
      useDismissProps: { enabled: false },
      focusManagerProps: { disabled: true },
      elementProps: {
        className: dragging
          ? "prompt-side-menu prompt-side-menu--dragging"
          : "prompt-side-menu",
        style: { zIndex: 20 },
      },
    }),
    [dragging, open, whileElementsMounted],
  );

  return (
    <GenericPopover reference={reference} {...floatingUIOptions}>
      {open ? (
        <PromptDragHandleMenuOpenProvider onOpenChange={setDragHandleMenuOpen}>
          <SideMenu>
            <PromptEditorAddBlockButton />
            <DragHandleButton dragHandleMenu={PromptEditorDragHandleMenu} />
          </SideMenu>
        </PromptDragHandleMenuOpenProvider>
      ) : null}
    </GenericPopover>
  );
}
