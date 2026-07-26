import { useMemo, useState } from "react";
import type {
  TaskTemplateDetail,
  TemplateFunctionBinding,
  TemplateFunctionInputDef,
  TemplateFunctionRef,
} from "@/types";
import { RepoScopePicker } from "./RepoScopePicker";

export type TemplateBindDraft = {
  templateId: string;
  name: string;
  worktreeId: string | null;
  inputs: TemplateFunctionInputDef[];
  bindings: Record<
    string,
    { paths: string[]; functions: TemplateFunctionRef[] }
  >;
};

type Props = {
  drafts: TemplateBindDraft[];
  pending: boolean;
  error: string | null;
  onChange: (next: TemplateBindDraft[]) => void;
  onCancel: () => void;
  onConfirm: () => void;
};

export function buildBindDraftsFromDetails(
  details: TaskTemplateDetail[],
  worktreeByTemplate: Record<string, string | null>,
): TemplateBindDraft[] {
  return details.map((d) => {
    const inputs = d.payload.function_inputs ?? [];
    const bindings: TemplateBindDraft["bindings"] = {};
    for (const input of inputs) {
      bindings[input.id] = { paths: [], functions: [] };
    }
    return {
      templateId: d.id,
      name: d.name,
      worktreeId: worktreeByTemplate[d.id] ?? d.payload.worktree_id ?? null,
      inputs,
      bindings,
    };
  });
}

export function bindingsFromDrafts(
  drafts: TemplateBindDraft[],
): Record<string, TemplateFunctionBinding[]> {
  const out: Record<string, TemplateFunctionBinding[]> = {};
  for (const draft of drafts) {
    out[draft.templateId] = draft.inputs.map((input) => {
      const val = draft.bindings[input.id] ?? { paths: [], functions: [] };
      if (input.kind === "function") {
        return { input_id: input.id, functions: val.functions };
      }
      return { input_id: input.id, paths: val.paths };
    });
  }
  return out;
}

export function validateBindDrafts(drafts: TemplateBindDraft[]): string | null {
  for (const draft of drafts) {
    for (const input of draft.inputs) {
      const required = input.required !== false;
      if (!required) continue;
      const val = draft.bindings[input.id] ?? { paths: [], functions: [] };
      if (input.kind === "function") {
        if (val.functions.length === 0) {
          return `${draft.name}: ${input.label || input.id} is required`;
        }
      } else if (val.paths.length === 0) {
        return `${draft.name}: ${input.label || input.id} is required`;
      }
    }
  }
  return null;
}

export function TemplateFunctionBindModal({
  drafts,
  pending,
  error,
  onChange,
  onCancel,
  onConfirm,
}: Props) {
  const [localError, setLocalError] = useState<string | null>(null);
  const title = useMemo(() => {
    if (drafts.length === 1) return `Inputs for ${drafts[0]!.name}`;
    return `Inputs for ${drafts.length} templates`;
  }, [drafts]);

  return (
    <div className="modal-backdrop" role="presentation" onClick={onCancel}>
      <div
        className="modal template-function-bind-modal"
        role="dialog"
        aria-modal="true"
        aria-labelledby="template-function-bind-title"
        onClick={(e) => e.stopPropagation()}
      >
        <header className="modal__header">
          <h2 id="template-function-bind-title">{title}</h2>
          <button type="button" className="secondary" onClick={onCancel} disabled={pending}>
            Close
          </button>
        </header>
        <div className="modal__body">
          <p className="hint">
            Choose the directories, files, or functions these template tasks may work in.
          </p>
          {drafts.map((draft, draftIndex) => (
            <section key={draft.templateId} className="template-function-bind-modal__block">
              {drafts.length > 1 ? <h3>{draft.name}</h3> : null}
              {draft.inputs.map((input) => {
                const val = draft.bindings[input.id] ?? { paths: [], functions: [] };
                return (
                  <div key={input.id} className="template-function-bind-modal__field">
                    <label className="field-label">
                      {input.label || input.id}
                      {input.required === false ? " (optional)" : ""}
                    </label>
                    <RepoScopePicker
                      kind={input.kind}
                      worktreeId={draft.worktreeId}
                      multiple={input.multiple === true}
                      paths={val.paths}
                      functions={val.functions}
                      disabled={pending}
                      onPathsChange={(paths) => {
                        const next = drafts.map((d, i) =>
                          i !== draftIndex
                            ? d
                            : {
                                ...d,
                                bindings: {
                                  ...d.bindings,
                                  [input.id]: { ...val, paths },
                                },
                              },
                        );
                        onChange(next);
                      }}
                      onFunctionsChange={(fns) => {
                        const next = drafts.map((d, i) =>
                          i !== draftIndex
                            ? d
                            : {
                                ...d,
                                bindings: {
                                  ...d.bindings,
                                  [input.id]: { ...val, functions: fns },
                                },
                              },
                        );
                        onChange(next);
                      }}
                    />
                  </div>
                );
              })}
            </section>
          ))}
          {localError || error ? (
            <div className="err" role="alert">
              <p>{localError || error}</p>
            </div>
          ) : null}
        </div>
        <footer className="modal__footer">
          <button type="button" className="secondary" onClick={onCancel} disabled={pending}>
            Cancel
          </button>
          <button
            type="button"
            disabled={pending}
            onClick={() => {
              const msg = validateBindDrafts(drafts);
              if (msg) {
                setLocalError(msg);
                return;
              }
              setLocalError(null);
              onConfirm();
            }}
          >
            Create tasks
          </button>
        </footer>
      </div>
    </div>
  );
}
