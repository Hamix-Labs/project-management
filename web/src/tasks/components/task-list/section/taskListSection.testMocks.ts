import { afterEach, beforeEach, vi } from "vitest";

const { mockPatchTask, mockDeleteTask } = vi.hoisted(() => ({
  mockPatchTask: vi.fn(),
  mockDeleteTask: vi.fn(),
}));

const isUiFeatureOmitted = vi.hoisted(() =>
  vi.fn<(feature: string) => boolean>(() => false),
);

vi.mock("@/launch/omittedFeatures", () => ({
  isUiFeatureOmitted: (feature: string) => isUiFeatureOmitted(feature),
}));

vi.mock("@/api", async (importOriginal) => {
  const actual = await importOriginal<typeof import("@/api")>();
  return {
    ...actual,
    patchTask: mockPatchTask,
    deleteTask: mockDeleteTask,
  };
});

beforeEach(() => {
  mockPatchTask.mockReset();
  mockDeleteTask.mockReset();
  mockDeleteTask.mockResolvedValue(undefined);
  isUiFeatureOmitted.mockImplementation(() => false);
});

afterEach(() => {
  vi.restoreAllMocks();
});
