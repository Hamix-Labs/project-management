import { useEffect } from "react";
import {
  consumePromptEditorReturn,
  writePromptEditorReturn,
} from "./promptEditorSession";

type ResumeHandlers = {
  setNewPrompt: (html: string) => void;
  setNewTitle: (title: string) => void;
  resumeComposeFromPromptEditor: () => void;
  promptEditorSuspended: boolean;
};

/**
 * After navigating back from PromptEditorPage, apply compose return payload once.
 * Polish returns are left for task-detail dialogs to consume.
 */
export function usePromptEditorReturnResume(handlers: ResumeHandlers): void {
  const {
    setNewPrompt,
    setNewTitle,
    resumeComposeFromPromptEditor,
    promptEditorSuspended,
  } = handlers;

  useEffect(() => {
    const payload = consumePromptEditorReturn();
    if (!payload) return;
    if (payload.resumePolish) {
      writePromptEditorReturn(payload);
      return;
    }
    if (!payload.resumeCompose) return;
    if (payload.html !== undefined) {
      setNewPrompt(payload.html);
    }
    if (payload.title !== undefined) {
      setNewTitle(payload.title);
    }
    if (promptEditorSuspended) {
      resumeComposeFromPromptEditor();
    }
  }, [
    promptEditorSuspended,
    resumeComposeFromPromptEditor,
    setNewPrompt,
    setNewTitle,
  ]);
}
