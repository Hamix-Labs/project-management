import { Link } from "react-router-dom";
import {
  BlockNotePromptEditor,
  PromptEditorDocHeader,
  PromptEditorTopbar,
} from "@/components/prompt-editor";
import { PromptEditorSessionAlert } from "@/components/prompt-editor/chrome/PromptEditorSessionAlert";
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
        title={c.title}
        saveStatus={c.saveStatus}
        saveErrorDetail={c.saveError?.detail}
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
          disabled={c.leavePending || !c.ready}
          onTitleCommit={c.onTitleCommit}
        />

        {c.hydrateWarning ? (
          <PromptEditorSessionAlert
            title={c.hydrateWarning.title}
            detail={c.hydrateWarning.detail}
            variant="warning"
            onDismiss={c.dismissHydrateWarning}
          />
        ) : null}

        {c.saveError ? (
          <PromptEditorSessionAlert
            title={c.saveError.title}
            detail={c.saveError.detail}
            onRetry={
              c.saveError.code === "rename_failed" ? undefined : c.retrySave
            }
            retryLabel="Retry save"
          />
        ) : null}

        <div className="prompt-editor-page__body">
          {c.status === "loading" ? (
            <div className="prompt-editor-page__skeleton" aria-busy="true">
              <p className="prompt-editor-page__loading">Loading prompt…</p>
              <div className="prompt-editor-page__skeleton-line prompt-editor-page__skeleton-line--title" />
              <div className="prompt-editor-page__skeleton-line" />
              <div className="prompt-editor-page__skeleton-line" />
              <div className="prompt-editor-page__skeleton-line prompt-editor-page__skeleton-line--short" />
            </div>
          ) : null}

          {c.loadError ? (
            <PromptEditorSessionAlert
              title={c.loadError.title}
              detail={c.loadError.detail}
              onRetry={c.retryLoad}
              onBack={c.leaveWithoutSave}
              retryLabel="Retry"
              backLabel="Back"
            />
          ) : null}

          {c.ready ? (
            <BlockNotePromptEditor
              key={`${c.sourceKind}-${c.sourceId}`}
              id="prompt-editor-page"
              initialHtml={c.html}
              onChange={c.onChange}
              onHydrateFallback={c.onHydrateFallback}
              disabled={c.leavePending}
              placeholder={
                c.launch?.placeholder ?? "Write the implementation brief…"
              }
              worktreeId={c.worktreeId}
              repositoryId={c.repositoryId}
            />
          ) : null}
        </div>
      </div>
    </div>
  );
}
