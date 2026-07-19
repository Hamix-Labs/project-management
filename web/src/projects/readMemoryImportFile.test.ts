import { describe, expect, it } from "vitest";
import { MAX_PROJECT_CONTEXT_BODY_BYTES } from "./projectContextLimits";
import {
  MemoryImportError,
  previewMemoryImportText,
  readMemoryImportFile,
  sanitizeMemoryAlias,
  validateMemoryAlias,
} from "./readMemoryImportFile";

function fileFrom(
  name: string,
  contents: string,
  options?: { sizeOverride?: number },
): File {
  const blob = new Blob([contents], { type: "text/plain" });
  const file = new File([blob], name, { type: "text/plain" });
  if (options?.sizeOverride !== undefined) {
    Object.defineProperty(file, "size", {
      value: options.sizeOverride,
    });
  }
  return file;
}

describe("readMemoryImportFile", () => {
  it("reads txt content and derives a sanitized alias", async () => {
    const out = await readMemoryImportFile(
      fileFrom("API Plan.txt", "Use REST for v1."),
    );
    expect(out.text).toBe("Use REST for v1.");
    expect(out.defaultAlias).toBe("API Plan");
  });

  it("strips UTF-8 BOM", async () => {
    const out = await readMemoryImportFile(
      fileFrom("notes.md", "\uFEFF# Heading\nBody"),
    );
    expect(out.text.startsWith("#")).toBe(true);
  });

  it("rejects disallowed extensions", async () => {
    await expect(
      readMemoryImportFile(fileFrom("secret.pdf", "%PDF")),
    ).rejects.toBeInstanceOf(MemoryImportError);
  });

  it("rejects oversize files before reading", async () => {
    await expect(
      readMemoryImportFile(
        fileFrom("big.md", "x", {
          sizeOverride: MAX_PROJECT_CONTEXT_BODY_BYTES + 1,
        }),
      ),
    ).rejects.toThrow(/512/);
  });

  it("rejects NUL bytes", async () => {
    await expect(
      readMemoryImportFile(fileFrom("bin.txt", "a\0b")),
    ).rejects.toThrow(/binary/i);
  });

  it("rejects empty files", async () => {
    await expect(
      readMemoryImportFile(fileFrom("empty.md", "   \n")),
    ).rejects.toThrow(/empty/i);
  });
});

describe("sanitizeMemoryAlias / validateMemoryAlias", () => {
  it("strips path separators and control characters", () => {
    expect(sanitizeMemoryAlias("../evil\nname")).toBe("evil name");
  });

  it("rejects empty and overlong aliases", () => {
    expect(validateMemoryAlias("   ")).toMatch(/required/i);
    expect(validateMemoryAlias("a".repeat(201))).toMatch(/200/);
    expect(validateMemoryAlias("Good alias")).toBeNull();
  });
});

describe("previewMemoryImportText", () => {
  it("truncates long previews without changing the source", () => {
    const text = Array.from({ length: 25 }, (_, i) => `line ${i}`).join("\n");
    const preview = previewMemoryImportText(text, 3);
    expect(preview.split("\n")).toHaveLength(4);
    expect(preview.endsWith("…")).toBe(true);
    expect(text.split("\n")).toHaveLength(25);
  });
});
