// @vitest-environment jsdom
import { describe, expect, it } from "vitest";
import { sanitizePromptHtml } from "@/lib/promptFormat";

describe("sanitizePromptHtml file embeds", () => {
  it("preserves file-embed data attributes", () => {
    const input = `<div class="file-embed" data-repo-file-embed="true" data-path="src/a.ts" data-line-start="1" data-line-end="3"><div class="file-embed-head"><span class="file-embed-path">src/a.ts</span><span class="file-embed-range">lines 1–3</span></div></div>`;
    const out = sanitizePromptHtml(input);
    expect(out).toContain('data-repo-file-embed="true"');
    expect(out).toContain('data-path="src/a.ts"');
    expect(out).toContain('data-line-start="1"');
    expect(out).toContain('data-line-end="3"');
    expect(out).toContain("file-embed-path");
  });

  it("still preserves repo file chips", () => {
    const input = `<p><span class="repo-file-chip" data-repo-file="true" data-path="x.go">@x.go</span></p>`;
    const out = sanitizePromptHtml(input);
    expect(out).toContain('data-repo-file="true"');
    expect(out).toContain('data-path="x.go"');
  });
});
