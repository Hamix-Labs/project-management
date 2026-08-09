import { createContext, useContext, type ReactNode } from "react";

/**
 * Lets {@link PromptEditorDragHandleMenu} report menu open/close without
 * forking BlockNote's `DragHandleButton`. The menu dropdown unmounts on hide
 * (`unmountOnHide`), so a mount/unmount beacon inside it is the open signal.
 *
 * Kept as context because `dragHandleMenu` is typed as `FC` with no props.
 */
const PromptDragHandleMenuOpenContext = createContext<
  ((open: boolean) => void) | null
>(null);

export function PromptDragHandleMenuOpenProvider({
  onOpenChange,
  children,
}: {
  onOpenChange: (open: boolean) => void;
  children: ReactNode;
}) {
  return (
    <PromptDragHandleMenuOpenContext.Provider value={onOpenChange}>
      {children}
    </PromptDragHandleMenuOpenContext.Provider>
  );
}

export function usePromptDragHandleMenuOpenChange():
  | ((open: boolean) => void)
  | null {
  return useContext(PromptDragHandleMenuOpenContext);
}
