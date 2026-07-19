import { useState, type FormEvent } from "react";
import type { ProjectContextRelation } from "@/types";
import type { ProjectContextMutations } from "./projectContextPanelHelpers";

export function useProjectContextFormState(
  mutations: ProjectContextMutations,
) {
  const [importOpen, setImportOpen] = useState(false);
  const [addEdgeOpen, setAddEdgeOpen] = useState(false);
  const [newEdgeSourceID, setNewEdgeSourceID] = useState("");
  const [newEdgeTargetID, setNewEdgeTargetID] = useState("");
  const [newEdgeRelation, setNewEdgeRelation] =
    useState<ProjectContextRelation>("related");
  const [newEdgeStrength, setNewEdgeStrength] = useState("3");
  const [newEdgeNote, setNewEdgeNote] = useState("");
  const [newEdgeEditorKey, setNewEdgeEditorKey] = useState(0);

  function submitImport(input: { title: string; body: string }) {
    mutations.createContextMutation.mutate(
      {
        kind: "note",
        title: input.title,
        body: input.body,
        pinned: false,
      },
      {
        onSuccess: () => {
          setImportOpen(false);
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
    importOpen,
    setImportOpen,
    addEdgeOpen,
    setAddEdgeOpen,
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
    submitImport,
    submitEdge,
    openAddEdge,
  };
}
