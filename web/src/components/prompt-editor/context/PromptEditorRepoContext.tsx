import { createContext, useContext } from "react";

export type PromptEditorRepoContextValue = {
  worktreeId?: string;
};

export const PromptEditorRepoContext = createContext<PromptEditorRepoContextValue>(
  {},
);

export function usePromptEditorRepo(): PromptEditorRepoContextValue {
  return useContext(PromptEditorRepoContext);
}
