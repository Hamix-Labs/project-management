import { describe, expect, it } from "vitest";
import {
  decideBlockTypeConversion,
  flattenBlockContentToPlainText,
  isConvertibleBlock,
  offeredBlockTypeTargets,
  targetMatchesBlock,
} from "./decideBlockTypeConversion";
import {
  PROMPT_BLOCK_TYPE_TARGETS,
  PROMPT_CODE_BLOCK_DEFAULT_LANGUAGE,
  promptBlockTypeTargetByKey,
} from "./promptBlockTypeTargets";

const paragraph = {
  type: "paragraph",
  props: {},
  content: [
    { type: "text", text: "Hello ", styles: { bold: true } },
    {
      type: "link",
      href: "https://example.com",
      content: [{ type: "text", text: "world", styles: {} }],
    },
  ],
};

const heading2 = {
  type: "heading",
  props: { level: 2, isToggleable: false },
  content: [{ type: "text", text: "Title", styles: { italic: true } }],
};

const codeBlock = {
  type: "codeBlock",
  props: { language: "python" },
  content: [{ type: "text", text: "print(1)", styles: {} }],
};

const withMention = {
  type: "paragraph",
  content: [
    { type: "text", text: "See ", styles: {} },
    {
      type: "repoFileMention",
      props: { path: "web/src/app.tsx", lineStart: "1", lineEnd: "10" },
    },
  ],
};

const withChildren = {
  type: "bulletListItem",
  props: {},
  content: [{ type: "text", text: "parent", styles: {} }],
  children: [
    {
      type: "paragraph",
      content: [{ type: "text", text: "child", styles: {} }],
    },
  ],
};

function target(key: string) {
  const found = promptBlockTypeTargetByKey(key);
  if (!found) {
    throw new Error(`missing target ${key}`);
  }
  return found;
}

describe("isConvertibleBlock", () => {
  it("accepts inline and plain content blocks", () => {
    expect(isConvertibleBlock(paragraph)).toBe(true);
    expect(isConvertibleBlock(codeBlock)).toBe(true);
  });

  it("rejects contentless and unknown structural blocks", () => {
    expect(isConvertibleBlock({ type: "repoFileEmbed" })).toBe(false);
    expect(isConvertibleBlock({ type: "table", content: { type: "tableContent", rows: [] } })).toBe(
      false,
    );
    expect(isConvertibleBlock(undefined)).toBe(false);
  });
});

describe("targetMatchesBlock / offeredBlockTypeTargets", () => {
  it("treats code-block language as irrelevant for matching", () => {
    expect(targetMatchesBlock(codeBlock, target("code_block"))).toBe(true);
  });

  it("distinguishes heading levels", () => {
    expect(targetMatchesBlock(heading2, target("heading_2"))).toBe(true);
    expect(targetMatchesBlock(heading2, target("heading"))).toBe(false);
  });

  it("offers the full catalog for convertible sources and nothing otherwise", () => {
    expect(offeredBlockTypeTargets(paragraph)).toEqual(PROMPT_BLOCK_TYPE_TARGETS);
    expect(offeredBlockTypeTargets({ type: "repoFileEmbed" })).toEqual([]);
  });
});

describe("flattenBlockContentToPlainText", () => {
  it("keeps marks' text, link text, and mention paths", () => {
    expect(flattenBlockContentToPlainText(paragraph.content)).toBe("Hello world");
    expect(flattenBlockContentToPlainText(withMention.content)).toBe(
      "See web/src/app.tsx",
    );
  });
});

describe("decideBlockTypeConversion", () => {
  it("is a no-op when the block already matches the target", () => {
    expect(decideBlockTypeConversion(paragraph, target("paragraph"))).toEqual({
      action: "noop",
    });
    expect(decideBlockTypeConversion(codeBlock, target("code_block"))).toEqual({
      action: "noop",
    });
  });

  it("is unavailable for non-convertible sources", () => {
    expect(
      decideBlockTypeConversion({ type: "repoFileEmbed" }, target("paragraph")),
    ).toEqual({ action: "unavailable" });
  });

  it("converts paragraph to heading without rewriting content", () => {
    expect(decideBlockTypeConversion(paragraph, target("heading_2"))).toEqual({
      action: "convert",
      liftChildren: false,
      update: {
        type: "heading",
        props: { level: 2, isToggleable: false },
      },
    });
  });

  it("converts heading to paragraph without rewriting content", () => {
    expect(decideBlockTypeConversion(heading2, target("paragraph"))).toEqual({
      action: "convert",
      liftChildren: false,
      update: { type: "paragraph", props: undefined },
    });
  });

  it("converts paragraph to each list flavour and quote", () => {
    for (const key of [
      "bullet_list",
      "numbered_list",
      "check_list",
      "toggle_list",
      "quote",
    ] as const) {
      const decision = decideBlockTypeConversion(paragraph, target(key));
      expect(decision.action).toBe("convert");
      if (decision.action === "convert") {
        expect(decision.update.content).toBeUndefined();
        expect(decision.update.type).toBe(target(key).type);
      }
    }
  });

  it("round-trips every catalog type against paragraph without content loss for inline pairs", () => {
    for (const t of PROMPT_BLOCK_TYPE_TARGETS) {
      if (t.key === "paragraph") {
        continue;
      }
      const forward = decideBlockTypeConversion(paragraph, t);
      expect(forward.action).toBe("convert");
      if (forward.action !== "convert") {
        continue;
      }
      if (t.contentKind === "inline") {
        expect(forward.update.content).toBeUndefined();
      } else {
        expect(forward.update.content).toEqual([
          { type: "text", text: "Hello world", styles: {} },
        ]);
      }

      const asTarget = {
        type: forward.update.type,
        props: forward.update.props ?? {},
        content:
          forward.update.content ??
          (t.contentKind === "plain"
            ? [{ type: "text", text: "Hello world", styles: {} }]
            : paragraph.content),
      };
      const back = decideBlockTypeConversion(asTarget, target("paragraph"));
      expect(back.action).toBe("convert");
      if (back.action === "convert" && t.contentKind === "plain") {
        expect(back.update.content).toEqual([
          { type: "text", text: "Hello world", styles: {} },
        ]);
      }
    }
  });

  it("strips marks when converting into a code block and applies the default language", () => {
    expect(decideBlockTypeConversion(paragraph, target("code_block"))).toEqual({
      action: "convert",
      liftChildren: false,
      update: {
        type: "codeBlock",
        props: { language: PROMPT_CODE_BLOCK_DEFAULT_LANGUAGE },
        content: [{ type: "text", text: "Hello world", styles: {} }],
      },
    });
  });

  it("keeps code text and drops language when converting out of a code block", () => {
    expect(decideBlockTypeConversion(codeBlock, target("paragraph"))).toEqual({
      action: "convert",
      liftChildren: false,
      update: {
        type: "paragraph",
        props: undefined,
        content: [{ type: "text", text: "print(1)", styles: {} }],
      },
    });
  });

  it("preserves children on convert (liftChildren stays false for catalog targets)", () => {
    const decision = decideBlockTypeConversion(withChildren, target("paragraph"));
    expect(decision).toEqual({
      action: "convert",
      liftChildren: false,
      update: { type: "paragraph", props: undefined },
    });
  });

  it("stays idempotent: convert then re-select the same target is a no-op", () => {
    const first = decideBlockTypeConversion(paragraph, target("heading_2"));
    expect(first.action).toBe("convert");
    if (first.action !== "convert") {
      return;
    }
    const after = {
      type: first.update.type,
      props: first.update.props,
      content: paragraph.content,
    };
    expect(decideBlockTypeConversion(after, target("heading_2"))).toEqual({
      action: "noop",
    });
  });
});
