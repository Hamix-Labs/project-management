import { useEffect, useMemo, useState } from "react";
import { useSearchParams } from "react-router-dom";
import { Button } from "@/components/ui";
import { useDocumentTitle } from "@/shared/useDocumentTitle";
import { EmptyState } from "@/shared/EmptyState";
import { useDelayedTrue } from "@/lib/useDelayedTrue";
import { useDebouncedTrimmedValue } from "@/hooks/useDebouncedTrimmedValue";
import { TASK_TIMINGS } from "@/constants/tasks";
import { TaskDraftsListSkeleton } from "@/components/skeletons/TaskDraftsListSkeleton";
import { useGlobalRepositories } from "./hooks/useGlobalRepositories";
import { useGlobalGitMutations } from "./hooks/useGlobalGitMutations";
import { RepositoriesListTable } from "./components/RepositoriesListTable";
import { RegisterRepositoryModal } from "./modals/RegisterRepositoryModal";
import {
  deriveWorktreesPageMode,
  worktreesPageErrorMessage,
  worktreesPageTitle,
} from "./worktreesPageMode";
import { repositoryMatchesSearchQuery } from "./repositoryDisplay";
import { worktreeGitCopy } from "./worktreeGitCopy";
import {
  WorktreesPlusIcon,
  WorktreesSearchIcon,
  WorktreesClearIcon,
} from "./components/WorktreesIcons";

export function RepositoriesListPage() {
  const repositoriesQuery = useGlobalRepositories();
  const mutations = useGlobalGitMutations();
  const [searchParams, setSearchParams] = useSearchParams();
  const [registerOpen, setRegisterOpen] = useState(false);
  const [searchInput, setSearchInput] = useState("");

  const repositories = repositoriesQuery.data ?? [];
  const debouncedQ = useDebouncedTrimmedValue(searchInput, 300);
  const filteredRepositories = useMemo(
    () => repositories.filter((repo) => repositoryMatchesSearchQuery(repo, debouncedQ)),
    [repositories, debouncedQ],
  );
  const pageMode = deriveWorktreesPageMode({
    isLoading: repositoriesQuery.isLoading && !repositoriesQuery.data,
    isError: repositoriesQuery.isError,
    repositoryCount: repositories.length,
  });
  const pageTitle = worktreesPageTitle();
  useDocumentTitle(pageTitle);
  const showSearch = pageMode === "setup" || pageMode === "manage";
  const showRegisterButton = pageMode === "setup" || pageMode === "manage";

  const showSkeleton = useDelayedTrue(
    pageMode === "loading",
    TASK_TIMINGS.draftResumeMinLoadingMs,
  );

  useEffect(() => {
    if (searchParams.get("register") !== "1") return;
    setRegisterOpen(true);
    setSearchParams({}, { replace: true });
  }, [searchParams, setSearchParams]);

  return (
    <div className="repositories-page">
      <section
        className="repositories-card worktrees-page"
        aria-labelledby="worktrees-heading"
      >
        <div className="repositories-card__header">
          <div className="repositories-card__header-text">
            <h1 id="worktrees-heading" className="repositories-card__title">
              {pageTitle}
            </h1>
            <p className="repositories-card__subtitle">
              {worktreeGitCopy.repositoriesPageSubtitle}
            </p>
          </div>
          {showRegisterButton ? (
            <Button
              type="button"
              variant="primary"
              className="repositories-card__register-btn"
              onClick={() => setRegisterOpen(true)}
            >
              <WorktreesPlusIcon className="repositories-card__register-icon" aria-hidden />
              {worktreeGitCopy.registerRepository}
            </Button>
          ) : null}
        </div>

        {showSearch ? (
          <div className="repositories-card__toolbar">
            <div className="repositories-card__search" role="search">
              <label htmlFor="repositories-search" className="visually-hidden">
                Search repositories
              </label>
              <WorktreesSearchIcon
                className="repositories-card__search-icon"
                aria-hidden
              />
              <input
                id="repositories-search"
                type="search"
                className="repositories-card__search-input"
                placeholder={worktreeGitCopy.searchRepositoriesPlaceholder}
                autoComplete="off"
                value={searchInput}
                onChange={(e) => setSearchInput(e.target.value)}
              />
              {searchInput.trim() ? (
                <button
                  type="button"
                  className="repositories-card__search-clear"
                  aria-label="Clear search field"
                  onClick={() => setSearchInput("")}
                >
                  <WorktreesClearIcon aria-hidden />
                </button>
              ) : null}
            </div>
            <p className="repositories-card__count">
              {worktreeGitCopy.repositoriesSearchCount(
                filteredRepositories.length,
                repositories.length,
              )}
            </p>
          </div>
        ) : null}

        {pageMode === "error" ? (
          <div className="repositories-card__error err" role="alert">
            <p>{worktreesPageErrorMessage(repositoriesQuery.error)}</p>
            <div className="task-detail-error-actions">
              <button
                type="button"
                className="secondary"
                onClick={() => {
                  void repositoriesQuery.refetch();
                }}
              >
                Try again
              </button>
            </div>
          </div>
        ) : null}

        <div className="repositories-card__body">
          {showSkeleton ? <TaskDraftsListSkeleton /> : null}
          {!showSkeleton ? (
            <>
              {pageMode === "setup" ? (
                <div className="repositories-card__empty-setup">
                  <EmptyState
                    title="Register a repository to get started"
                    description="Hamix needs a git checkout before you can register worktrees, bind branches, and run agent tasks."
                    hideIcon
                    className="empty-state--in-table empty-state--task-list-fresh"
                  />
                </div>
              ) : null}
              {pageMode === "manage" && filteredRepositories.length === 0 && debouncedQ ? (
                <div className="repositories-card__empty-search">
                  <span className="repositories-card__empty-search-icon-wrap">
                    <WorktreesSearchIcon aria-hidden />
                  </span>
                  <div className="repositories-card__empty-search-text">
                    <p className="repositories-card__empty-search-title">
                      {worktreeGitCopy.repositoriesSearchEmptyTitle}
                    </p>
                    <p className="repositories-card__empty-search-description">
                      {worktreeGitCopy.repositoriesSearchEmptyDescription(debouncedQ)}
                    </p>
                  </div>
                  <Button
                    type="button"
                    variant="secondary"
                    className="repositories-card__empty-search-clear"
                    onClick={() => setSearchInput("")}
                  >
                    {worktreeGitCopy.clearSearch}
                  </Button>
                </div>
              ) : null}
              {pageMode === "manage" && filteredRepositories.length > 0 ? (
                <RepositoriesListTable repositories={filteredRepositories} />
              ) : null}
            </>
          ) : null}
        </div>

        <RegisterRepositoryModal
          open={registerOpen}
          pending={mutations.createRepository.isPending}
          error={mutations.createRepository.error}
          onClose={() => {
            setRegisterOpen(false);
            mutations.createRepository.reset();
          }}
          onSubmit={(input) => {
            void mutations.createRepository
              .mutateAsync(input)
              .then(() => setRegisterOpen(false));
          }}
        />
      </section>
    </div>
  );
}

