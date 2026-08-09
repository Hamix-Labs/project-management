import { act, render, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { PROMPT_BLOCK_ACTIVE_ATTR } from "./promptBlockElement";

type PopoverReference = { getBoundingClientRect: () => DOMRect };
type PopoverProps = {
  reference?: PopoverReference;
  elementProps?: { className?: string };
  useFloatingOptions?: {
    open?: boolean;
    whileElementsMounted?: (
      reference: unknown,
      floating: unknown,
      update: () => void,
    ) => void | (() => void);
  };
  children?: React.ReactNode;
};

let popoverProps: PopoverProps | null = null;
let popoverMounts = 0;
let editorChangeCallbacks: Array<() => void> = [];
let editorDom: HTMLElement | null = null;
let sideMenuState: { show: boolean; block?: { id: string } } | undefined;
let selectionBlockIds: string[] | undefined;
const floatingUpdate = vi.fn();
const getSelection = vi.fn(() =>
  selectionBlockIds
    ? { blocks: selectionBlockIds.map((id) => ({ id })) }
    : undefined,
);

vi.mock("@blocknote/core/extensions", () => ({
  SideMenuExtension: { key: "sideMenu" },
}));

vi.mock("./PromptEditorAddBlockButton", () => ({
  PromptEditorAddBlockButton: () => <div data-testid="add-block" />,
}));

vi.mock("./PromptEditorDragHandleMenu", async () => {
  const React = await vi.importActual<typeof import("react")>("react");
  const mod = await vi.importActual<
    typeof import("./promptDragHandleMenuOpenContext")
  >("./promptDragHandleMenuOpenContext");

  return {
    PromptEditorDragHandleMenu: function MockDragHandleMenu() {
      const onOpenChange = mod.usePromptDragHandleMenuOpenChange();
      React.useEffect(() => {
        onOpenChange?.(true);
        return () => onOpenChange?.(false);
      }, [onOpenChange]);
      return React.createElement("div", { "data-testid": "drag-handle-menu" });
    },
  };
});

vi.mock("@blocknote/react", async () => {
  const React = await vi.importActual<typeof import("react")>("react");

  return {
    GenericPopover: (props: PopoverProps) => {
      popoverProps = props;
      React.useEffect(() => {
        popoverMounts += 1;
        return props.useFloatingOptions?.whileElementsMounted?.(
          document.createElement("div"),
          document.createElement("div"),
          floatingUpdate,
        );
        // Mount-only, mirroring FloatingUI's whileElementsMounted lifecycle.
        // eslint-disable-next-line react-hooks/exhaustive-deps
      }, []);
      return (
        <div data-testid="popover" className={props.elementProps?.className}>
          {props.children}
        </div>
      );
    },
    SideMenu: ({ children }: { children?: React.ReactNode }) => (
      <div data-testid="side-menu">{children}</div>
    ),
    DragHandleButton: ({
      dragHandleMenu: Menu,
    }: {
      dragHandleMenu?: React.ComponentType;
    }) => {
      const [menuOpen, setMenuOpen] = React.useState(false);
      return (
        <div>
          <button
            type="button"
            data-testid="drag-handle"
            onClick={() => setMenuOpen((open) => !open)}
          />
          {menuOpen && Menu ? <Menu /> : null}
        </div>
      );
    },
    useBlockNoteEditor: () => ({ getSelection }),
    useEditorChange: (callback: () => void) => {
      React.useEffect(() => {
        editorChangeCallbacks.push(callback);
        return () => {
          editorChangeCallbacks = editorChangeCallbacks.filter(
            (entry) => entry !== callback,
          );
        };
      }, [callback]);
    },
    useEditorDOMElement: () => editorDom,
    useExtensionState: (
      _extension: unknown,
      ctx: { selector: (state: unknown) => unknown },
    ) => ctx.selector(sideMenuState),
  };
});

import { PromptEditorSideMenu } from "./PromptEditorSideMenu";

function fireEditorChange() {
  act(() => {
    for (const callback of [...editorChangeCallbacks]) {
      callback();
    }
  });
}

function stubRect(element: Element, rect: DOMRectInit) {
  element.getBoundingClientRect = () => DOMRect.fromRect(rect);
}

function buildEditorDom() {
  const dom = document.createElement("div");
  const blockGroup = document.createElement("div");
  stubRect(blockGroup, { x: 40, y: 0, width: 600, height: 400 });
  dom.appendChild(blockGroup);
  return { dom, blockGroup };
}

function appendBlock(blockGroup: Element, blockId: string, y: number) {
  const outer = document.createElement("div");
  outer.setAttribute("data-node-type", "blockOuter");
  outer.setAttribute("data-id", blockId);
  const container = document.createElement("div");
  container.setAttribute("data-node-type", "blockContainer");
  container.setAttribute("data-id", blockId);
  stubRect(container, { x: 60, y, width: 560, height: 28 });
  outer.appendChild(container);
  blockGroup.appendChild(outer);
  return { outer, container };
}

function mountHost() {
  const host = document.createElement("div");
  document.body.appendChild(host);
  return host;
}

function dispatch(target: EventTarget, type: "dragstart" | "drop") {
  act(() => {
    target.dispatchEvent(new Event(type, { bubbles: true }));
  });
}

beforeEach(() => {
  popoverProps = null;
  popoverMounts = 0;
  editorChangeCallbacks = [];
  floatingUpdate.mockClear();
  getSelection.mockClear();
  selectionBlockIds = undefined;
  sideMenuState = { show: true, block: { id: "block-1" } };
  const built = buildEditorDom();
  editorDom = built.dom;
  appendBlock(built.blockGroup, "block-1", 120);
});

describe("PromptEditorSideMenu", () => {
  it("re-measures the anchor after the block's node is replaced", () => {
    render(<PromptEditorSideMenu editorHost={mountHost()} />);

    const reference = popoverProps?.reference;
    expect(reference?.getBoundingClientRect().y).toBe(120);

    const blockGroup = editorDom!.firstElementChild!;
    blockGroup.querySelector('[data-node-type="blockOuter"]')!.remove();
    appendBlock(blockGroup, "block-1", 220);

    // Same reference object, new rect — no re-render or remount involved.
    expect(popoverProps?.reference).toBe(reference);
    expect(reference?.getBoundingClientRect().y).toBe(220);
  });

  it("leaves the drag source untouched during the dragstart handler", () => {
    const host = mountHost();
    const view = render(<PromptEditorSideMenu editorHost={host} />);

    dispatch(host, "dragstart");

    // Chrome cancels the drag if the source is restyled from within dragstart.
    expect(view.getByTestId("popover").className).toBe("prompt-side-menu");
  });

  it("suppresses the menu during a drag without unmounting the drag handle", async () => {
    const host = mountHost();
    const view = render(<PromptEditorSideMenu editorHost={host} />);

    expect(view.getByTestId("popover").className).toBe("prompt-side-menu");

    dispatch(host, "dragstart");
    await waitFor(() =>
      expect(view.getByTestId("popover").className).toContain(
        "prompt-side-menu--dragging",
      ),
    );

    // The handle is the drag source; unmounting it would strand the drag.
    expect(view.getByTestId("side-menu")).toBeTruthy();

    dispatch(host, "drop");

    expect(view.getByTestId("popover").className).toBe("prompt-side-menu");
  });

  it("never remounts the popover across a full drag cycle", async () => {
    const host = mountHost();
    const view = render(<PromptEditorSideMenu editorHost={host} />);

    expect(popoverMounts).toBe(1);

    dispatch(host, "dragstart");
    await waitFor(() =>
      expect(view.getByTestId("popover").className).toContain(
        "prompt-side-menu--dragging",
      ),
    );
    dispatch(host, "drop");

    expect(popoverMounts).toBe(1);
  });

  it("repositions instead of remounting when the document changes", async () => {
    render(<PromptEditorSideMenu editorHost={mountHost()} />);
    await waitFor(() => expect(floatingUpdate).toHaveBeenCalled());
    floatingUpdate.mockClear();

    fireEditorChange();
    fireEditorChange();

    // Deferred out of the change handler and coalesced into one pass.
    expect(floatingUpdate).not.toHaveBeenCalled();
    await waitFor(() => expect(floatingUpdate).toHaveBeenCalledTimes(1));
    expect(popoverMounts).toBe(1);
  });

  it("repositions once a drag finishes, since the block moved under it", async () => {
    const host = mountHost();
    const view = render(<PromptEditorSideMenu editorHost={host} />);
    await waitFor(() => expect(floatingUpdate).toHaveBeenCalled());

    dispatch(host, "dragstart");
    await waitFor(() =>
      expect(view.getByTestId("popover").className).toContain(
        "prompt-side-menu--dragging",
      ),
    );
    floatingUpdate.mockClear();

    dispatch(host, "drop");

    await waitFor(() => expect(floatingUpdate).toHaveBeenCalledTimes(1));
  });

  it("renders the prompt add-block button next to BlockNote's drag handle", () => {
    const view = render(<PromptEditorSideMenu editorHost={mountHost()} />);

    expect(view.getByTestId("add-block")).toBeTruthy();
    expect(view.getByTestId("drag-handle")).toBeTruthy();
  });

  it("stays closed while no block is hovered", () => {
    sideMenuState = { show: false };

    const view = render(<PromptEditorSideMenu editorHost={mountHost()} />);

    expect(view.queryByTestId("side-menu")).toBeNull();
    expect(popoverProps?.useFloatingOptions?.open).toBe(false);
    expect(popoverProps?.reference).toBeUndefined();
  });

  it("does not highlight a block on side-menu hover alone", () => {
    render(<PromptEditorSideMenu editorHost={mountHost()} />);

    const block = editorDom!.querySelector(
      '[data-node-type="blockContainer"][data-id="block-1"]',
    );
    expect(block?.hasAttribute(PROMPT_BLOCK_ACTIVE_ATTR)).toBe(false);
  });

  it("highlights the target block while the drag-handle menu is open", async () => {
    const view = render(<PromptEditorSideMenu editorHost={mountHost()} />);

    act(() => {
      view.getByTestId("drag-handle").click();
    });

    await waitFor(() => {
      const block = editorDom!.querySelector(
        '[data-node-type="blockContainer"][data-id="block-1"]',
      );
      expect(block?.getAttribute(PROMPT_BLOCK_ACTIVE_ATTR)).toBe("true");
    });
    expect(view.getByTestId("drag-handle-menu")).toBeTruthy();
  });

  it("clears the highlight when the drag-handle menu closes", async () => {
    const view = render(<PromptEditorSideMenu editorHost={mountHost()} />);
    const block = () =>
      editorDom!.querySelector(
        '[data-node-type="blockContainer"][data-id="block-1"]',
      );

    act(() => {
      view.getByTestId("drag-handle").click();
    });
    await waitFor(() =>
      expect(block()?.getAttribute(PROMPT_BLOCK_ACTIVE_ATTR)).toBe("true"),
    );

    act(() => {
      view.getByTestId("drag-handle").click();
    });

    await waitFor(() =>
      expect(block()?.hasAttribute(PROMPT_BLOCK_ACTIVE_ATTR)).toBe(false),
    );
  });

  it("clears the highlight when the side menu dismisses", async () => {
    const host = mountHost();
    const view = render(<PromptEditorSideMenu editorHost={host} />);
    const block = () =>
      editorDom!.querySelector(
        '[data-node-type="blockContainer"][data-id="block-1"]',
      );

    act(() => {
      view.getByTestId("drag-handle").click();
    });
    await waitFor(() =>
      expect(block()?.getAttribute(PROMPT_BLOCK_ACTIVE_ATTR)).toBe("true"),
    );

    sideMenuState = { show: false };
    view.rerender(<PromptEditorSideMenu editorHost={host} />);

    await waitFor(() =>
      expect(block()?.hasAttribute(PROMPT_BLOCK_ACTIVE_ATTR)).toBe(false),
    );
  });

  it("highlights every selected block when the handle's block is in the selection", async () => {
    const blockGroup = editorDom!.firstElementChild!;
    appendBlock(blockGroup, "block-2", 160);
    selectionBlockIds = ["block-1", "block-2"];

    const view = render(<PromptEditorSideMenu editorHost={mountHost()} />);

    act(() => {
      view.getByTestId("drag-handle").click();
    });

    await waitFor(() => {
      expect(
        editorDom!
          .querySelector(
            '[data-node-type="blockContainer"][data-id="block-1"]',
          )
          ?.getAttribute(PROMPT_BLOCK_ACTIVE_ATTR),
      ).toBe("true");
      expect(
        editorDom!
          .querySelector(
            '[data-node-type="blockContainer"][data-id="block-2"]',
          )
          ?.getAttribute(PROMPT_BLOCK_ACTIVE_ATTR),
      ).toBe("true");
    });
  });

  it("highlights the dragged block while a drag is in flight", async () => {
    const host = mountHost();
    render(<PromptEditorSideMenu editorHost={host} />);
    const block = () =>
      editorDom!.querySelector(
        '[data-node-type="blockContainer"][data-id="block-1"]',
      );

    dispatch(host, "dragstart");

    await waitFor(() =>
      expect(block()?.getAttribute(PROMPT_BLOCK_ACTIVE_ATTR)).toBe("true"),
    );

    dispatch(host, "drop");

    await waitFor(() =>
      expect(block()?.hasAttribute(PROMPT_BLOCK_ACTIVE_ATTR)).toBe(false),
    );
  });
});
