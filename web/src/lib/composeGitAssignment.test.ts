import { describe, expect, it } from "vitest";
import {
  applyRepoScopedDefaults,
  decideComposeGitAssignment,
  hydrateAssignmentFromPayload,
  initFreshAssignment,
  isFreshAssignment,
  selectRepository,
  selectWorktree,
} from "./composeGitAssignment";

const REPO_A = "00000000-0000-4000-8000-000000000010";
const REPO_B = "00000000-0000-4000-8000-000000000011";
const PROJ_A = "00000000-0000-4000-8000-000000000040";
const PROJ_B = "00000000-0000-4000-8000-000000000041";
const WT_FEATURE = "00000000-0000-4000-8000-000000000021";

const projects = [
  { id: PROJ_A, is_default: true, status: "active" as const },
  { id: PROJ_B, is_default: false, status: "active" as const },
];

const settledQueries = {
  repositories: [{ id: REPO_A }],
  repositoriesLoading: false,
  projects,
  projectsLoading: false,
  worktreesLoading: false,
};

describe("composeGitAssignment", () => {
  it("initFresh selects the only repository when assignment is empty", () => {
    const next = initFreshAssignment(
      { repositoryId: "", projectId: "", worktreeId: "" },
      [{ id: REPO_A }],
    );
    expect(next.repositoryId).toBe(REPO_A);
  });

  it("initFresh does not overwrite a hydrated repository", () => {
    const current = { repositoryId: REPO_B, projectId: PROJ_A, worktreeId: WT_FEATURE };
    const next = initFreshAssignment(current, [{ id: REPO_A }]);
    expect(next).toEqual(current);
  });

  it("hydrate preserves saved assignment ids", () => {
    expect(
      hydrateAssignmentFromPayload({
        repositoryId: REPO_A,
        projectId: PROJ_A,
        worktreeId: WT_FEATURE,
      }),
    ).toEqual({
      repositoryId: REPO_A,
      projectId: PROJ_A,
      worktreeId: WT_FEATURE,
    });
  });

  it("selectRepository clears project and worktree", () => {
    expect(
      selectRepository(
        { repositoryId: REPO_A, projectId: PROJ_A, worktreeId: WT_FEATURE },
        REPO_B,
      ),
    ).toEqual({ repositoryId: REPO_B, projectId: "", worktreeId: "" });
  });

  it("applyRepoScopedDefaults clears worktree (server allocates)", () => {
    const current = { repositoryId: REPO_A, projectId: PROJ_A, worktreeId: WT_FEATURE };
    expect(applyRepoScopedDefaults(current, projects)).toEqual({
      repositoryId: REPO_A,
      projectId: PROJ_A,
      worktreeId: "",
    });
  });

  it("applyRepoScopedDefaults picks default project when empty", () => {
    const current = { repositoryId: REPO_A, projectId: "", worktreeId: "" };
    expect(applyRepoScopedDefaults(current, projects).projectId).toBe(PROJ_A);
  });

  it("isFreshAssignment detects an untouched compose form", () => {
    expect(isFreshAssignment({ repositoryId: "", projectId: "", worktreeId: "" })).toBe(true);
    expect(isFreshAssignment({ repositoryId: REPO_A, projectId: "", worktreeId: "" })).toBe(false);
  });

  it("selectWorktree sets worktree only", () => {
    expect(
      selectWorktree(
        { repositoryId: REPO_A, projectId: PROJ_A, worktreeId: "" },
        WT_FEATURE,
      ).worktreeId,
    ).toBe(WT_FEATURE);
  });

  it("decide applies sole-repo default on a fresh assignment", () => {
    const next = decideComposeGitAssignment(
      { repositoryId: "", projectId: "", worktreeId: "" },
      settledQueries,
      { projectChosenByUser: false },
    );
    expect(next.repositoryId).toBe(REPO_A);
  });

  it("decide fills default project when user has not chosen one", () => {
    const next = decideComposeGitAssignment(
      { repositoryId: REPO_A, projectId: "", worktreeId: "" },
      settledQueries,
      { projectChosenByUser: false },
    );
    expect(next.projectId).toBe(PROJ_A);
  });

  it("decide keeps user project pick over late defaults", () => {
    const next = decideComposeGitAssignment(
      { repositoryId: REPO_A, projectId: PROJ_B, worktreeId: "" },
      settledQueries,
      { projectChosenByUser: true },
    );
    expect(next.projectId).toBe(PROJ_B);
  });

  it("decide waits while scoped lists are loading", () => {
    const current = { repositoryId: REPO_A, projectId: "", worktreeId: "" };
    const next = decideComposeGitAssignment(
      current,
      { ...settledQueries, projectsLoading: true },
      { projectChosenByUser: false },
    );
    expect(next).toEqual(current);
  });
});
