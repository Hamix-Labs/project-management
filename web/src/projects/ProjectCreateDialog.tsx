import { useEffect, useId, useRef, useState, type FormEvent } from "react";
import { CustomSelect } from "@/components/custom-select";
import { Button } from "@/components/ui";
import { useGlobalRepositories } from "@/hooks/useGlobalRepositories";
import { Modal } from "@/shared/Modal";
import { MutationErrorBanner } from "@/shared/MutationErrorBanner";

type Props = {
  saving: boolean;
  error?: unknown;
  repositoryId?: string;
  onCancel: () => void;
  onSubmit: (input: { name: string; description: string; repository_id: string }) => void;
};

export function ProjectCreateDialog({
  saving,
  error = null,
  repositoryId: fixedRepositoryId,
  onCancel,
  onSubmit,
}: Props) {
  const titleId = useId();
  const nameId = useId();
  const descriptionId = useId();
  const nameRef = useRef<HTMLInputElement>(null);

  const [name, setName] = useState("");
  const [description, setDescription] = useState("");
  const [selectedRepositoryId, setSelectedRepositoryId] = useState(fixedRepositoryId ?? "");

  const repositoriesQuery = useGlobalRepositories({
    enabled: fixedRepositoryId == null,
  });
  const repositories = repositoriesQuery.data ?? [];

  useEffect(() => {
    nameRef.current?.focus();
  }, []);

  useEffect(() => {
    if (fixedRepositoryId) {
      setSelectedRepositoryId(fixedRepositoryId);
      return;
    }
    if (repositories.length === 1 && selectedRepositoryId === "") {
      setSelectedRepositoryId(repositories[0]!.id);
    }
  }, [fixedRepositoryId, repositories, selectedRepositoryId]);

  const trimmedName = name.trim();
  const canSubmit =
    trimmedName.length > 0 && selectedRepositoryId.trim() !== "" && !saving;

  function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!canSubmit) return;
    onSubmit({
      name: trimmedName,
      description: description.trim(),
      repository_id: selectedRepositoryId.trim(),
    });
  }

  const repoOptions = repositories.map((repo) => ({
    value: repo.id,
    label: repo.path,
  }));

  return (
    <Modal
      onClose={onCancel}
      labelledBy={titleId}
      busy={saving}
      busyLabel="Creating project…"
      dismissibleWhileBusy
    >
      <section className="panel modal-sheet pl__create-dialog">
        <header className="pl__create-dialog-head">
          <h2 id={titleId}>New project</h2>
          <p className="pl__create-dialog-help">
            Additional projects group tasks within a repository. The system default
            is created automatically when you register a repo.
          </p>
        </header>

        <form
          className="pl__create-dialog-form"
          onSubmit={handleSubmit}
          aria-label="Create project"
        >
          {fixedRepositoryId == null ? (
            <CustomSelect
              id={`${titleId}-repository`}
              label="Repository"
              value={selectedRepositoryId}
              options={repoOptions}
              onChange={setSelectedRepositoryId}
              disabled={saving || repositoriesQuery.isLoading}
              requirement="required"
            />
          ) : null}

          <div className="field">
            <label htmlFor={nameId}>Name</label>
            <input
              ref={nameRef}
              id={nameId}
              type="text"
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="e.g. Billing platform"
              required
              disabled={saving}
              autoComplete="off"
            />
          </div>

          <div className="field">
            <label htmlFor={descriptionId}>
              Description <span className="pl__create-dialog-optional">— optional</span>
            </label>
            <textarea
              id={descriptionId}
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              placeholder="One line of context that helps your team and agents understand what this project is for."
              rows={3}
              disabled={saving}
            />
          </div>

          <MutationErrorBanner
            error={error}
            fallback="Could not create project."
            className="pl__create-dialog-err"
          />

          <div className="row stack-row-actions pl__create-dialog-actions">
            <Button
              type="button"
              variant="secondary"
              onClick={onCancel}
              disabled={saving}
            >
              Cancel
            </Button>
            <Button type="submit" variant="primary" disabled={!canSubmit} loading={saving}>
              Create project
            </Button>
          </div>
        </form>
      </section>
    </Modal>
  );
}
