import { Link } from "react-router-dom";
import {
  BlockNotePromptEditor,
  PromptEditorDocHeader,
  PromptEditorTopbar,
} from "@/components/prompt-editor";
import { MutationErrorBanner } from "@/shared/MutationErrorBanner";
import { useDocumentTitle } from "@/shared/useDocumentTitle";
import { usePromptEditorPageController } from "../prompt-editor/usePromptEditorPageController";

export function PromptEditorPage() {
  const c = usePromptEditorPageController();
  useDocumentTitle(c.kindOk ? `${c.title} · Prompt` : "Prompt Editor");

  if (!c.kindOk || !c.sourceId) {
    return (
      <div className="prompt-editor-page">
        <div className="prompt-editor-page__canvas">
          <p>Unknown prompt document.</p>
          <Link to="/">Back</Link>
        </div>
      </div>
    );
  }

  return (
    <div className="prompt-editor-page">
      <PromptEditorTopbar
        crumbKindLabel={c.crumbKindLabel}
        title={c.title}
        saveStatus={c.saveStatus}
        leavePending={c.leavePending}
        onBack={() => void c.leaveEditor()}
        onRetrySave={c.retrySave}
      />

      <div className="prompt-editor-page__canvas">
        <PromptEditorDocHeader
          title={c.title}
          editedLabel={c.editedLabel}
          wordCountLabel={c.wordCountLabel}
          repoLabel={c.repoLabel}
        />

        <MutationErrorBanner error={c.loadError} />
        <MutationErrorBanner error={c.saveError} />

        <div className="prompt-editor-page__body">
          {c.loaded ? (
            <BlockNotePromptEditor
              id="prompt-editor-page"
              value={c.html}
              onChange={c.onChange}
              disabled={c.leavePending}
              placeholder={
                c.launch?.placeholder ?? "Write the implementation brief…"
              }
              worktreeId={c.worktreeId}
            />
          ) : (
            <div className="prompt-editor-page__skeleton" aria-busy="true">
              <div className="prompt-editor-page__skeleton-line prompt-editor-page__skeleton-line--title" />
              <div className="prompt-editor-page__skeleton-line" />
              <div className="prompt-editor-page__skeleton-line" />
              <div className="prompt-editor-page__skeleton-line prompt-editor-page__skeleton-line--short" />
              <span className="visually-hidden">Loading prompt…</span>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
