import { useCallback, useEffect, useMemo, useRef } from "react";
import { SideMenuExtension } from "@blocknote/core/extensions";
import {
  GenericPopover,
  SideMenu,
  useEditorChange,
  useEditorDOMElement,
  useExtensionState,
  type GenericPopoverReference,
} from "@blocknote/react";
import { promptSideMenuAnchorRect } from "./promptSideMenuAnchor";
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
 */
export function PromptEditorSideMenu({
  editorHost,
}: {
  editorHost: HTMLElement | null;
}) {
  const editorDom = useEditorDOMElement();
  const dragging = usePromptBlockDragState(editorHost);
  const { show, blockId } = useExtensionState(SideMenuExtension, {
    selector: (state) => ({
      show: state?.show ?? false,
      blockId: state?.block?.id as string | undefined,
    }),
  });

  const open = show && blockId !== undefined;

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
      {open ? <SideMenu /> : null}
    </GenericPopover>
  );
}
