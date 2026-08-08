import { useEffect, useState } from "react";
import { SideMenuExtension } from "@blocknote/core/extensions";
import {
  SideMenuController,
  useExtensionState,
} from "@blocknote/react";

type PromptSideMenuState = {
  show: boolean;
  referencePos: DOMRect;
  block?: { id?: string };
};

export function promptSideMenuPositionKey(
  state: PromptSideMenuState | undefined,
) {
  if (!state?.show || !state.block?.id) {
    return "hidden";
  }

  const { referencePos } = state;
  return [
    state.block.id,
    referencePos.x,
    referencePos.y,
    referencePos.width,
    referencePos.height,
  ]
    .map((value) =>
      typeof value === "number" ? Math.round(value).toString() : value,
    )
    .join(":");
}

function isBlockNoteDrag(event: DragEvent, editorHost: HTMLElement) {
  if (event.target instanceof Node && editorHost.contains(event.target)) {
    return true;
  }

  return Array.from(event.dataTransfer?.types ?? []).includes(
    "blocknote/html",
  );
}

export function PromptEditorSideMenuController({
  editorHost,
}: {
  editorHost: HTMLElement | null;
}) {
  const positionKey = useExtensionState(SideMenuExtension, {
    selector: promptSideMenuPositionKey,
  });
  const [placementRevision, setPlacementRevision] = useState(0);

  useEffect(() => {
    if (!editorHost) {
      return;
    }

    const ownerWindow = editorHost.ownerDocument.defaultView ?? window;
    let blockDragActive = false;
    let frame = 0;

    const scheduleReanchor = () => {
      if (frame) {
        return;
      }

      frame = ownerWindow.requestAnimationFrame(() => {
        frame = 0;
        setPlacementRevision((revision) => revision + 1);
      });
    };

    const onDragStart = (event: DragEvent) => {
      if (!isBlockNoteDrag(event, editorHost)) {
        return;
      }

      blockDragActive = true;
      scheduleReanchor();
    };

    const onDragMove = (event: DragEvent) => {
      if (!blockDragActive || !isBlockNoteDrag(event, editorHost)) {
        return;
      }

      scheduleReanchor();
    };

    const onDragFinish = (event: DragEvent) => {
      if (!blockDragActive && !isBlockNoteDrag(event, editorHost)) {
        return;
      }

      blockDragActive = false;
      scheduleReanchor();
    };

    const { ownerDocument } = editorHost;
    ownerDocument.addEventListener("dragstart", onDragStart);
    ownerDocument.addEventListener("dragover", onDragMove);
    ownerDocument.addEventListener("drop", onDragFinish);
    ownerDocument.addEventListener("dragend", onDragFinish);

    return () => {
      if (frame) {
        ownerWindow.cancelAnimationFrame(frame);
      }
      ownerDocument.removeEventListener("dragstart", onDragStart);
      ownerDocument.removeEventListener("dragover", onDragMove);
      ownerDocument.removeEventListener("drop", onDragFinish);
      ownerDocument.removeEventListener("dragend", onDragFinish);
    };
  }, [editorHost]);

  return (
    <SideMenuController key={`${positionKey}:${placementRevision}`} />
  );
}
