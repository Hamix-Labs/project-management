import type { PromptDocumentAdapter, PromptDocumentRef } from "./types";

/**
 * Extension seams for future Prompt Editor capabilities.
 * Do not implement features here yet — keep adapters swappable.
 */
export type PromptEditorCapability =
  | "templates"
  | "reusableSections"
  | "embeddedMedia"
  | "aiAssist"
  | "versioning";

export type PromptDocumentStore = {
  resolve: (ref: PromptDocumentRef) => PromptDocumentAdapter;
};

/** Marker type for optional version snapshots (future). */
export type PromptDocumentVersion = {
  id: string;
  createdAt: string;
  html: string;
  label?: string;
};
