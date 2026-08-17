import { describe, expect, it } from "vitest";
import {
  buildHamixSlashMenuConfig,
  HAMIX_SLASH_ENABLED_ITEMS,
} from "./hamixSlashMenu";

describe("hamixSlashMenu", () => {
  it("keeps Notion blocks and omits Cloud AI, images, TOC, and user mention", () => {
    expect(HAMIX_SLASH_ENABLED_ITEMS).toContain("heading_1");
    expect(HAMIX_SLASH_ENABLED_ITEMS).toContain("divider");
    expect(HAMIX_SLASH_ENABLED_ITEMS).not.toContain("image");
    expect(HAMIX_SLASH_ENABLED_ITEMS).not.toContain("mention");
    expect(HAMIX_SLASH_ENABLED_ITEMS).not.toContain("toc");
    expect(HAMIX_SLASH_ENABLED_ITEMS).not.toContain("continue_writing");
    expect(HAMIX_SLASH_ENABLED_ITEMS).not.toContain("ai_ask_button");
  });

  it("appends Hamix Ask AI and mention-a-file after the template rows", () => {
    const titles = (buildHamixSlashMenuConfig(undefined).customItems ?? []).map(
      (item) => item.title,
    );
    expect(titles).toEqual(["Ask AI", "Mention a file"]);
  });
});
