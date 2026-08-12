import { useState } from "react";
import { Navigate, useParams } from "react-router-dom";
import { AppErrorBoundary } from "@/shared/AppErrorBoundary";
import { useTasksAppContext } from "../app/TasksAppProvider";
import type { ComposeMode } from "./composeMode";
import { TaskComposePageBody } from "./TaskComposePageBody";

const COMPOSE_FALLBACK =
  "Something went wrong while opening the task form.";

/**
 * Routed compose page for create/edit task and template (ADR-0100).
 * Seeds the shared create-flow hooks from the URL, then renders page chrome.
 */
export function TaskComposePage({ mode }: { mode: ComposeMode }) {
  const [layerKey, setLayerKey] = useState(0);
  const app = useTasksAppContext();
  return (
    <AppErrorBoundary
      variant="modal-layer"
      fallbackMessage={COMPOSE_FALLBACK}
      onRecover={() => {
        app.closeEdit();
        setLayerKey((k) => k + 1);
      }}
    >
      <TaskComposePageBody key={layerKey} mode={mode} />
    </AppErrorBoundary>
  );
}

export function TaskComposeNewPage() {
  return <TaskComposePage mode={{ kind: "task-create" }} />;
}

export function TaskComposeEditPage() {
  const { taskId } = useParams();
  if (!taskId) return <Navigate to="/" replace />;
  return <TaskComposePage mode={{ kind: "task-edit", taskId }} />;
}

export function TemplateComposeNewPage() {
  return <TaskComposePage mode={{ kind: "template-create" }} />;
}

export function TemplateComposeEditPage() {
  const { templateId } = useParams();
  if (!templateId) return <Navigate to="/templates" replace />;
  return <TaskComposePage mode={{ kind: "template-edit", templateId }} />;
}
