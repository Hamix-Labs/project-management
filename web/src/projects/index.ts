// ProjectListPage / ProjectDetailPage are NOT re-exported here on purpose —
// they are route-level entry points that App.tsx loads via React.lazy().
// Re-exporting them from the barrel would force Rollup to bundle them into
// whichever chunk imports the barrel, defeating the code-split. Import them
// directly from "@/projects/<PageName>" in tests or lazy-loaders only.
export { ProjectSelect } from "./ProjectSelect";
export { projectQueryKeys } from "./queryKeys";
export { useProject, useProjects } from "./hooks";
