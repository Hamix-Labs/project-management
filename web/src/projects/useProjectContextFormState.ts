import { useState, type FormEvent } from "react";
import { promptHasVisibleContent } from "@/lib/promptFormat";
import type { ProjectContextKind, ProjectContextRelation } from "@/types";
import type { ContextView, ProjectContextMutations } from "./projectContextPanelHelpers";

export function useProjectContextFormState(
  mutations: ProjectContextMutations,
) {
  const [contextView, setContextView] = useState<ContextView>("list");
  const [addNodeOpen, setAddNodeOpen] = useState(false);
  const [addEdgeOpen, setAddEdgeOpen] = useState(false);
  const [newNodeBody, setNewNodeBody] = useState("");
  const [newNodeEditorKey, setNewNodeEditorKey] = useState(0);
  const [newEdgeSourceID, setNewEdgeSourceID] = useState("");
  const [newEdgeTargetID, setNewEdgeTargetID] = useState("");
  const [newEdgeRelation, setNewEdgeRelation] =
    useState<ProjectContextRelation>("related");
  const [newEdgeStrength, setNewEdgeStrength] = useState("3");
  const [newEdgeNote, setNewEdgeNote] = useState("");
  const [newEdgeEditorKey, setNewEdgeEditorKey] = useState(0);

  function submitContext(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const form = new FormData(event.currentTarget);
    const title = String(form.get("title") ?? "").trim();
    const body = newNodeBody.trim();
    if (!title || !promptHasVisibleContent(body)) return;
    const formEl = event.currentTarget;
    mutations.createContextMutation.mutate(
      {
        kind: String(form.get("kind") ?? "note") as ProjectContextKind,
        title,
        body,
        pinned: false,
      },
      {
        onSuccess: () => {
          formEl.reset();
          setNewNodeBody("");
          setNewNodeEditorKey((value) => value + 1);
          setAddNodeOpen(false);
        },
      },
    );
  }

  function submitEdge(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (
      !newEdgeSourceID ||
      !newEdgeTargetID ||
      newEdgeSourceID === newEdgeTargetID
    ) {
      return;
    }
    const formEl = event.currentTarget;
    mutations.createEdgeMutation.mutate(
      {
        source_context_id: newEdgeSourceID,
        target_context_id: newEdgeTargetID,
        relation: newEdgeRelation,
        strength: Number(newEdgeStrength),
        note: newEdgeNote.trim(),
      },
      {
        onSuccess: () => {
          formEl.reset();
          setNewEdgeSourceID("");
          setNewEdgeTargetID("");
          setNewEdgeRelation("related");
          setNewEdgeStrength("3");
          setNewEdgeNote("");
          setNewEdgeEditorKey((value) => value + 1);
          setAddEdgeOpen(false);
        },
      },
    );
  }

  function openAddEdge(sourceId = "") {
    setNewEdgeSourceID(sourceId);
    setNewEdgeTargetID("");
    setNewEdgeRelation("related");
    setNewEdgeStrength("3");
    setNewEdgeNote("");
    setNewEdgeEditorKey((value) => value + 1);
    setAddEdgeOpen(true);
  }

  return {
    contextView,
    setContextView,
    addNodeOpen,
    setAddNodeOpen,
    addEdgeOpen,
    setAddEdgeOpen,
    newNodeBody,
    setNewNodeBody,
    newNodeEditorKey,
    setNewNodeEditorKey,
    newEdgeSourceID,
    setNewEdgeSourceID,
    newEdgeTargetID,
    setNewEdgeTargetID,
    newEdgeRelation,
    setNewEdgeRelation,
    newEdgeStrength,
    setNewEdgeStrength,
    newEdgeNote,
    setNewEdgeNote,
    newEdgeEditorKey,
    setNewEdgeEditorKey,
    submitContext,
    submitEdge,
    openAddEdge,
  };
}
