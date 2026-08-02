import { Link } from "react-router-dom";
import { BlockNotePromptEditor } from "@/components/prompt-editor";
import { MutationErrorBanner } from "@/shared/MutationErrorBanner";
import { usePromptEditorPageController } from "../prompt-editor/usePromptEditorPageController";

export function PromptEditorPage() {
  const c = usePromptEditorPageController();

  if (!c.kindOk || !c.sourceId) {
    return (
      <div className="prompt-editor-page">
        <p>Unknown prompt document.</p>
        <Link to="/">Back</Link>
      </div>
    );
  }

  return (
    <div className="prompt-editor-page">
      <header className="prompt-editor-page__header">
        <div className="prompt-editor-page__titles">
          <p className="prompt-editor-page__eyebrow">Prompt Editor</p>
          <h1 className="prompt-editor-page__title">
            {c.launch?.title ?? "Prompt"}
          </h1>
        </div>
        <div className="prompt-editor-page__actions">
          <span className="prompt-editor-page__status" aria-live="polite">
            {c.saving ? "Saving…" : c.saveError ? "Save failed" : ""}
          </span>
          <button
            type="button"
            className="primary"
            disabled={c.donePending || !c.loaded}
            onClick={() => void c.onDone()}
          >
            {c.donePending ? "Saving…" : "Done"}
          </button>
        </div>
      </header>

      <MutationErrorBanner error={c.loadError} />
      <MutationErrorBanner error={c.saveError} />

      <div className="prompt-editor-page__body task-create-editor-shell">
        {c.loaded ? (
          <BlockNotePromptEditor
            id="prompt-editor-page"
            value={c.html}
            onChange={c.onChange}
            disabled={c.donePending}
            placeholder={
              c.launch?.placeholder ?? "Write the implementation brief…"
            }
            worktreeId={c.launch?.worktreeId}
          />
        ) : (
          <p className="prompt-editor-page__loading">Loading prompt…</p>
        )}
      </div>
    </div>
  );
}
