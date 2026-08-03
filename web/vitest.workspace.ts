import path from "node:path";
import { fileURLToPath } from "node:url";
import react from "@vitejs/plugin-react";
import { defineWorkspace } from "vitest/config";

const __dirname = path.dirname(fileURLToPath(import.meta.url));

const appIncludes = ["src/app/**/*.test.tsx"];
const taskPagesIncludes = ["src/tasks/pages/**/*.test.tsx"];
const taskCreateIncludes = [
  "src/tasks/create/**/*.test.tsx",
  "!src/tasks/create/hooks/**",
];
const settingsIncludes = ["src/settings/SettingsPage.test.tsx"];
const projectsIncludes = [
  "src/projects/ProjectListPage.test.tsx",
  "src/projects/ProjectDetailPage.test.tsx",
];
const worktreesIncludes = [
  "src/worktrees/RepositoriesListPage.test.tsx",
];

const fullAppIncludes = [
  ...appIncludes,
  ...taskPagesIncludes,
  ...taskCreateIncludes,
  ...settingsIncludes,
  ...projectsIncludes,
  ...worktreesIncludes,
];

const sharedTest = {
  environment: "jsdom" as const,
  setupFiles: ["./src/test/setup.ts"],
  restoreMocks: true,
  // MSW patches global fetch in beforeAll; unstubbing globals after each test
  // drops interception and leaves integration tests waiting on real network timeouts.
  unstubGlobals: false,
};

const sharedVite = {
  plugins: [react()],
  resolve: {
    alias: {
      "@": path.resolve(__dirname, "src"),
    },
  },
};

/** Strict MSW for full-app verticals after F-09-04/F-09-05 spy migration. */
const harnessStrictTest = {
  ...sharedTest,
  testTimeout: 15_000,
  env: { HAMIX_MSW_UNHANDLED: "error" },
};

export default defineWorkspace([
  {
    ...sharedVite,
    test: {
      name: "unit",
      include: ["src/**/*.test.ts"],
      // node by default — avoid jsdom+MSW per-file tax. DOM/hook files set
      // // @vitest-environment jsdom at the top of the file.
      environment: "node",
      setupFiles: ["./src/test/setup.unit.ts"],
      restoreMocks: true,
      unstubGlobals: false,
    },
  },
  {
    ...sharedVite,
    test: {
      ...sharedTest,
      name: "components",
      include: ["src/**/*.test.tsx"],
      exclude: fullAppIncludes,
    },
  },
  {
    ...sharedVite,
    test: {
      ...harnessStrictTest,
      name: "app",
      include: appIncludes,
    },
  },
  {
    ...sharedVite,
    test: {
      ...harnessStrictTest,
      name: "task-pages",
      include: taskPagesIncludes,
    },
  },
  {
    ...sharedVite,
    test: {
      ...harnessStrictTest,
      name: "task-create",
      include: taskCreateIncludes,
    },
  },
  {
    ...sharedVite,
    test: {
      ...harnessStrictTest,
      name: "settings",
      include: settingsIncludes,
    },
  },
  {
    ...sharedVite,
    test: {
      ...harnessStrictTest,
      name: "projects",
      include: projectsIncludes,
    },
  },
  {
    ...sharedVite,
    test: {
      ...harnessStrictTest,
      name: "worktrees",
      include: worktreesIncludes,
    },
  },
]);
