import { afterEach, beforeEach, vi } from "vitest";
import { makeTask } from "@/test/taskDefaults";

const { mockPatchTask, mockCloseTask } = vi.hoisted(() => ({
  mockPatchTask: vi.fn(),
  mockCloseTask: vi.fn(),
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
    closeTask: mockCloseTask,
  };
});

beforeEach(() => {
  mockPatchTask.mockReset();
  mockCloseTask.mockReset();
  mockCloseTask.mockResolvedValue(
    makeTask({ id: "closed", status: "closed" }),
  );
  isUiFeatureOmitted.mockImplementation(() => false);
});

afterEach(() => {
  vi.restoreAllMocks();
});
