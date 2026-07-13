import { useEffect } from "react";

type ResettableMutation = {
  isIdle: boolean;
  reset: () => void;
};

export function useTaskCreateMutationResets(input: {
  createModalOpen: boolean;
  createMutation: ResettableMutation;
  saveDraftMutation: ResettableMutation;
  saveTemplateMutation: ResettableMutation;
  patchTemplateMutation: ResettableMutation;
}) {
  useEffect(() => {
    if (!input.createModalOpen && !input.saveDraftMutation.isIdle) {
      input.saveDraftMutation.reset();
    }
  }, [input.createModalOpen, input.saveDraftMutation]);

  useEffect(() => {
    if (!input.createModalOpen && !input.createMutation.isIdle) {
      input.createMutation.reset();
    }
  }, [input.createModalOpen, input.createMutation]);

  useEffect(() => {
    if (!input.createModalOpen && !input.saveTemplateMutation.isIdle) {
      input.saveTemplateMutation.reset();
    }
  }, [input.createModalOpen, input.saveTemplateMutation]);

  useEffect(() => {
    if (!input.createModalOpen && !input.patchTemplateMutation.isIdle) {
      input.patchTemplateMutation.reset();
    }
  }, [input.createModalOpen, input.patchTemplateMutation]);
}
