import { repositoryPathsEquivalent } from "./repositoryDisplay";

export type RegisteredRepositoryPaths = {
  path: string;
  host_path: string;
};

/** True when `candidatePath` matches a listed repository's path or host_path. */
export function isRepositoryAlreadyRegistered(
  candidatePath: string,
  registered: readonly RegisteredRepositoryPaths[],
): boolean {
  const trimmed = candidatePath.trim();
  if (trimmed === "") return false;
  return registered.some(
    (repo) =>
      repositoryPathsEquivalent(trimmed, repo.path) ||
      (repo.host_path.trim() !== "" &&
        repositoryPathsEquivalent(trimmed, repo.host_path)),
  );
}
