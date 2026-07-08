import { makeMutationTestWrapper } from "@/test/reactQuery";
import {
  mockAdd,
  mockDelete,
  mockPatch,
  mockPatchVerify,
} from "./useTaskDetailChecklist.testMocks";

export const TASK_A = "11111111-1111-4111-8111-111111111111";
export const TASK_B = "22222222-2222-4222-8222-222222222222";
export const ITEM_ID = "33333333-3333-4333-8333-333333333333";

export function setupChecklistTest() {
  return makeMutationTestWrapper();
}

export function resetChecklistMocks() {
  mockAdd.mockReset();
  mockPatch.mockReset();
  mockPatchVerify.mockReset();
  mockDelete.mockReset();
  mockAdd.mockResolvedValue({
    id: ITEM_ID,
    task_id: TASK_A,
    text: "criterion",
    done: false,
  });
  mockPatch.mockResolvedValue({
    id: ITEM_ID,
    task_id: TASK_A,
    text: "updated",
    done: false,
  });
  mockPatchVerify.mockResolvedValue({
    items: [{ id: ITEM_ID, sort_order: 0, text: "criterion", done: false }],
  });
  mockDelete.mockResolvedValue(undefined);
}
