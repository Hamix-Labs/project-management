import { useMemo, useRef } from "react";
import { useParams } from "react-router-dom";
import {
  createPromptDocumentAdapter,
  isPromptSourceKind,
} from "./promptDocumentAdapter";
import { readPromptEditorLaunch } from "./promptEditorSession";
import type { PromptEditorLaunchContext } from "./types";

/** Route params + launch context + document adapter for the Prompt Editor page. */
export function usePromptEditorRouteAdapter() {
  const { sourceKind = "", sourceId = "" } = useParams<{
    sourceKind: string;
    sourceId: string;
  }>();
  const launchRef = useRef<PromptEditorLaunchContext | null>(null);
  if (launchRef.current === null) {
    launchRef.current = readPromptEditorLaunch();
  }
  const launch = launchRef.current;
  const kindOk = isPromptSourceKind(sourceKind);
  const adapter = useMemo(() => {
    if (!kindOk || !sourceId) return null;
    return createPromptDocumentAdapter(
      { kind: sourceKind, id: sourceId },
      launch ?? undefined,
    );
  }, [kindOk, sourceKind, sourceId, launch]);

  return { sourceKind, sourceId, kindOk, launch, adapter };
}
