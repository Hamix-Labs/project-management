import { useEffect, useState, type FormEvent } from "react";
import type { Project } from "@/types";
import { usePatchProjectMutation } from "./mutations";

type Props = {
  project: Project;
};

export function ProjectSettingsPanel({ project }: Props) {
  const isDefaultProject = project.is_default;
  const [name, setName] = useState(project.name);
  const [description, setDescription] = useState(project.description ?? "");

  const patchProjectMutation = usePatchProjectMutation(project);

  useEffect(() => {
    setName(project.name);
    setDescription(project.description ?? "");
  }, [project.id, project.name, project.description, project.updated_at]);

  function submitProject(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (isDefaultProject) return;
    patchProjectMutation.mutate({
      name: name.trim(),
      description: description.trim(),
    });
  }

  function cancelEdits() {
    if (patchProjectMutation.isPending) return;
    setName(project.name);
    setDescription(project.description ?? "");
    patchProjectMutation.reset();
  }

  return (
    <section className="pd__card" aria-labelledby="pd-settings-title">
      <div className="pd__card-head">
        <h2 id="pd-settings-title" className="pd__card-title">
          Project settings
        </h2>
        <p className="pd__card-desc">Manage the core details for this project</p>
      </div>

      {isDefaultProject ? (
        <p className="pd__note">
          The default project is built in — its name is fixed.
        </p>
      ) : null}

      <form className="pd__settings-form" onSubmit={submitProject}>
        <div className="field">
          <label htmlFor="project-edit-name">Name</label>
          <input
            id="project-edit-name"
            name="name"
            value={name}
            onChange={(e) => setName(e.target.value)}
            required
            disabled={isDefaultProject}
            autoComplete="off"
          />
        </div>

        <div className="field">
          <label htmlFor="project-edit-description">
            Description{" "}
            <span className="pd__settings-form-optional">— optional</span>
          </label>
          <textarea
            id="project-edit-description"
            name="description"
            value={description}
            onChange={(e) => setDescription(e.target.value)}
            placeholder="One line of context that helps your team and agents understand what this project is for."
            rows={3}
            disabled={isDefaultProject}
          />
        </div>

        <div className="pd__settings-form-actions">
          <button
            type="button"
            className="secondary"
            disabled={isDefaultProject || patchProjectMutation.isPending}
            onClick={cancelEdits}
          >
            Cancel
          </button>
          <button
            type="submit"
            disabled={isDefaultProject || patchProjectMutation.isPending}
          >
            {patchProjectMutation.isPending ? "Saving…" : "Save changes"}
          </button>
        </div>
      </form>

      {patchProjectMutation.error ? (
        <div className="pd__inline-error" role="alert">
          {patchProjectMutation.error.message}
        </div>
      ) : null}
    </section>
  );
}
