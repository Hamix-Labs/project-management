export { ProjectContextEntryCard } from "./ProjectContextEntryCard";
// ProjectListPage / ProjectDetailPage / ProjectContextPage are NOT
// re-exported here on purpose — they are route-level entry points
// that App.tsx loads via React.lazy(). Re-exporting them from the
// barrel would force Rollup to bundle them into whichever chunk
// imports the barrel, defeating the code-split. Import them directly
// from "@/projects/<PageName>" in tests or lazy-loaders only.
export { ProjectSelect } from "./ProjectSelect";
export { ProjectContextPicker } from "./ProjectContextPicker";
export {
  MAX_SELECTED_PROJECT_CONTEXT_ITEMS,
  PROJECT_CONTEXT_SHORT_ID_LENGTH,
  mergeProjectContextSelection,
  projectContextShortId,
  selectedProjectContextItems,
} from "@/lib/projectContextRefs";
export { projectQueryKeys } from "./queryKeys";
export {
  useProject,
  useProjectContext,
  useProjects,
} from "./hooks";
export {
  useProjectContextPromptBinding,
  type UseProjectContextPromptBindingOptions,
} from "./useProjectContextPromptBinding";
