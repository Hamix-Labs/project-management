import type { Root } from "react-dom/client";
import { CodeLanguageToolbar } from "./CodeLanguageToolbar";
import {
  blockIdOf,
  findSelect,
  hideNativeSelect,
  positionMount,
  resolveCreateRoot,
} from "./codeToolbarDom";
import type { PromptCodeLanguage } from "./promptCodeBlockOptions";

/**
 * Structural view of the BlockNote editor. A code block's language lives in its
 * block props, so the toolbar reads and writes it there rather than poking the
 * node view's <select>, which BlockNote destroys and recreates on every update.
 */
export type CodeBlockLanguageEditor = {
  getBlock(id: string): { props?: { language?: string } } | undefined;
  updateBlock(id: string, update: { props: { language: string } }): unknown;
  onChange(callback: () => void): unknown;
};

type MountRecord = {
  root: Root;
  mountEl: HTMLElement;
  block: HTMLElement;
  blockId: string | null;
  select: HTMLSelectElement;
  onSelectChange: () => void;
  renderedValue: string | null;
  renderedDisabled: boolean | null;
};

export type CodeToolbarMountsOptions = {
  host: HTMLElement;
  editor?: CodeBlockLanguageEditor | null;
  languages: () => PromptCodeLanguage[];
  isDisabled: () => boolean;
  keyOf: (block: HTMLElement) => string;
};

/** Owns the portaled toolbar per code block: mount, render, reposition, teardown. */
export class CodeToolbarMounts {
  private readonly mounts = new Map<string, MountRecord>();

  constructor(private readonly options: CodeToolbarMountsOptions) {}

  /** Mount or refresh the toolbar for a code block; returns its mount key. */
  ensure(block: Element): string | null {
    if (!(block instanceof HTMLElement)) return null;
    const select = findSelect(block);
    if (!select) return null;

    const key = this.options.keyOf(block);
    const existing = this.mounts.get(key);
    if (existing) {
      existing.block = block;
      existing.blockId = blockIdOf(block);
      this.bindSelect(existing, select);
      this.render(existing);
      positionMount(this.options.host, block, existing.mountEl);
      return key;
    }
    return this.create(key, block, select);
  }

  /** Tear down toolbars whose blocks are no longer in the document. */
  prune(live: Set<string>) {
    for (const [key, record] of [...this.mounts]) {
      if (live.has(key)) continue;
      this.teardown(key, record);
    }
    this.options.host
      .querySelectorAll(":scope > .prompt-code-toolbar-root")
      .forEach((el) => {
        const kept = [...this.mounts.values()].some((r) => r.mountEl === el);
        if (!kept) el.remove();
      });
  }

  renderAll() {
    for (const record of this.mounts.values()) this.render(record);
  }

  repositionAll() {
    for (const record of this.mounts.values()) {
      if (!record.block.isConnected) continue;
      positionMount(this.options.host, record.block, record.mountEl);
    }
  }

  clear() {
    for (const [key, record] of [...this.mounts]) this.teardown(key, record);
    this.mounts.clear();
  }

  private create(
    key: string,
    block: HTMLElement,
    select: HTMLSelectElement,
  ): string | null {
    const mount = document.createElement("div");
    mount.className = "prompt-code-toolbar-root";
    this.options.host.appendChild(mount);
    positionMount(this.options.host, block, mount);

    block.querySelector("pre")?.setAttribute("spellcheck", "false");
    block.querySelector("code")?.setAttribute("spellcheck", "false");

    try {
      const record: MountRecord = {
        root: resolveCreateRoot()(mount),
        mountEl: mount,
        block,
        blockId: blockIdOf(block),
        select,
        onSelectChange: () => {
          const current = this.mounts.get(key);
          if (current) this.render(current, current.select.value);
        },
        renderedValue: null,
        renderedDisabled: null,
      };
      select.addEventListener("change", record.onSelectChange);
      hideNativeSelect(select);
      this.mounts.set(key, record);
      this.render(record);
      return key;
    } catch (err) {
      mount.remove();
      console.error("prompt-editor: code toolbar mount failed", err);
      return null;
    }
  }

  private teardown(key: string, record: MountRecord) {
    record.select.removeEventListener("change", record.onSelectChange);
    this.mounts.delete(key);
    record.mountEl.remove();
    const root = record.root;
    queueMicrotask(() => {
      try {
        root.unmount();
      } catch {
        // Root may already be gone after a fast remount.
      }
    });
  }

  private bindSelect(record: MountRecord, select: HTMLSelectElement) {
    if (record.select !== select) {
      record.select.removeEventListener("change", record.onSelectChange);
      record.select = select;
      select.addEventListener("change", record.onSelectChange);
    }
    hideNativeSelect(select);
  }

  private languageOf(record: MountRecord) {
    const { editor } = this.options;
    if (editor && record.blockId) {
      const language = editor.getBlock(record.blockId)?.props?.language;
      if (typeof language === "string" && language) return language;
    }
    return record.select.value;
  }

  private applyLanguage(record: MountRecord, languageId: string) {
    const { editor } = this.options;
    if (editor && record.blockId) {
      editor.updateBlock(record.blockId, { props: { language: languageId } });
      // The node view rerenders asynchronously; show the choice immediately.
      this.render(record, languageId);
      return;
    }
    record.select.value = languageId;
    record.select.dispatchEvent(new Event("change", { bubbles: true }));
  }

  private render(record: MountRecord, valueOverride?: string) {
    const value = valueOverride ?? this.languageOf(record);
    const disabled = this.options.isDisabled() || record.select.disabled;
    if (record.renderedValue === value && record.renderedDisabled === disabled) {
      return;
    }
    record.renderedValue = value;
    record.renderedDisabled = disabled;
    record.root.render(
      <CodeLanguageToolbar
        languages={this.options.languages()}
        value={value}
        disabled={disabled}
        onChange={(languageId) => this.applyLanguage(record, languageId)}
        onCopy={async () => {
          const text = record.block.querySelector("code")?.textContent ?? "";
          await navigator.clipboard.writeText(text);
        }}
      />,
    );
  }
}
