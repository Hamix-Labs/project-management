import {
  MAX_PROJECT_CONTEXT_BODY_BYTES,
  MAX_PROJECT_CONTEXT_TITLE_CHARS,
  MEMORY_IMPORT_EXTENSIONS,
} from "./projectContextLimits";

export type MemoryImportFileResult = {
  text: string;
  defaultAlias: string;
  fileName: string;
  byteSize: number;
};

export class MemoryImportError extends Error {
  constructor(message: string) {
    super(message);
    this.name = "MemoryImportError";
  }
}

function extensionOf(fileName: string): string {
  const base = fileName.split(/[/\\]/).pop() ?? fileName;
  const dot = base.lastIndexOf(".");
  if (dot <= 0) return "";
  return base.slice(dot).toLowerCase();
}

function basenameWithoutExtension(fileName: string): string {
  const base = fileName.split(/[/\\]/).pop() ?? fileName;
  const dot = base.lastIndexOf(".");
  if (dot <= 0) return base;
  return base.slice(0, dot);
}

/** Sanitize a user-facing alias derived from a filename or typed title. */
export function sanitizeMemoryAlias(raw: string): string {
  const withoutControls = [...raw]
    .map((ch) => {
      const code = ch.codePointAt(0) ?? 0;
      if (code < 32 || code === 127) return " ";
      if (ch === "/" || ch === "\\" || ch === "\0") return " ";
      return ch;
    })
    .join("")
    .replace(/\.\./g, " ");
  return withoutControls.replace(/\s+/g, " ").trim();
}

export function validateMemoryAlias(alias: string): string | null {
  const trimmed = sanitizeMemoryAlias(alias);
  if (!trimmed) return "Alias is required.";
  if ([...trimmed].length > MAX_PROJECT_CONTEXT_TITLE_CHARS) {
    return `Alias must be ${MAX_PROJECT_CONTEXT_TITLE_CHARS} characters or fewer.`;
  }
  return null;
}

function utf8ByteLength(text: string): number {
  return new TextEncoder().encode(text).length;
}

function assertAllowedExtension(fileName: string): void {
  const ext = extensionOf(fileName);
  if (
    !(MEMORY_IMPORT_EXTENSIONS as readonly string[]).includes(ext)
  ) {
    throw new MemoryImportError(
      "Only .txt and .md files can be imported.",
    );
  }
}

function normalizeImportedText(raw: string): string {
  let text = raw;
  if (text.charCodeAt(0) === 0xfeff) {
    text = text.slice(1);
  }
  if (text.includes("\0")) {
    throw new MemoryImportError(
      "File looks like binary content. Choose a plain text .txt or .md file.",
    );
  }
  text = text.trim();
  if (!text) {
    throw new MemoryImportError("File is empty.");
  }
  const bytes = utf8ByteLength(text);
  if (bytes > MAX_PROJECT_CONTEXT_BODY_BYTES) {
    throw new MemoryImportError(
      `Memory body must be ${formatMemoryImportBytes(MAX_PROJECT_CONTEXT_BODY_BYTES)} or smaller.`,
    );
  }
  return text;
}

async function readFileAsText(file: File): Promise<string> {
  if (typeof file.text === "function") {
    return file.text();
  }
  return await new Promise((resolve, reject) => {
    const reader = new FileReader();
    reader.onload = () => resolve(String(reader.result ?? ""));
    reader.onerror = () =>
      reject(reader.error ?? new Error("Could not read file"));
    reader.readAsText(file);
  });
}

/**
 * Validate and read a local memory import file as UTF-8 text.
 * Checks size before reading when possible to avoid loading huge files.
 */
export async function readMemoryImportFile(
  file: File,
): Promise<MemoryImportFileResult> {
  assertAllowedExtension(file.name);
  if (file.size > MAX_PROJECT_CONTEXT_BODY_BYTES) {
    throw new MemoryImportError(
      `Memory body must be ${formatMemoryImportBytes(MAX_PROJECT_CONTEXT_BODY_BYTES)} or smaller.`,
    );
  }

  const raw = await readFileAsText(file);
  const text = normalizeImportedText(raw);
  const defaultAlias = sanitizeMemoryAlias(
    basenameWithoutExtension(file.name),
  );
  if (!defaultAlias) {
    throw new MemoryImportError(
      "Could not derive an alias from the file name. Enter one manually.",
    );
  }

  return {
    text,
    defaultAlias,
    fileName: file.name.split(/[/\\]/).pop() ?? file.name,
    byteSize: utf8ByteLength(text),
  };
}

export function formatMemoryImportBytes(bytes: number): string {
  if (bytes < 1024) return `${bytes} B`;
  const kib = bytes / 1024;
  if (kib < 1024) return `${kib < 10 ? kib.toFixed(1) : Math.round(kib)} KiB`;
  return `${(kib / 1024).toFixed(1)} MiB`;
}

export function previewMemoryImportText(text: string, maxLines = 20): string {
  const lines = text.split(/\r?\n/);
  if (lines.length <= maxLines) return text;
  return `${lines.slice(0, maxLines).join("\n")}\n…`;
}
