import { vi } from "vitest";

const mocks = vi.hoisted(() => ({
  mockAdd: vi.fn(),
  mockPatch: vi.fn(),
  mockPatchVerify: vi.fn(),
  mockDelete: vi.fn(),
}));

vi.mock("@/api", async (importOriginal) => {
  const actual = await importOriginal<typeof import("@/api")>();
  return {
    ...actual,
    addChecklistItem: mocks.mockAdd,
    patchChecklistItemText: mocks.mockPatch,
    patchChecklistItemVerifyCommands: mocks.mockPatchVerify,
    deleteChecklistItem: mocks.mockDelete,
  };
});

export const mockAdd = mocks.mockAdd;
export const mockPatch = mocks.mockPatch;
export const mockPatchVerify = mocks.mockPatchVerify;
export const mockDelete = mocks.mockDelete;
