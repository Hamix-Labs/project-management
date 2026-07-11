import { useId } from "react";
import { FieldLabel } from "@/shared/FieldLabel";
import { TaskCreateDependsOnBrowseModal } from "./TaskCreateDependsOnBrowseModal";
import { TaskCreateDependsOnSearchRow } from "./TaskCreateDependsOnSearchRow";
import { TaskCreateDependsOnSelectedChips } from "./TaskCreateDependsOnSelectedChips";
import { TaskCreateDependsOnTypeaheadList } from "./TaskCreateDependsOnTypeaheadList";
import { useTaskCreateDependsOnPickerState } from "./useTaskCreateDependsOnPickerState";

type Props = {
  /**
   * Project the new task is being scoped to. The picker only surfaces
   * tasks that share this `project_id` — picking dependencies across
   * project boundaries is not a use case we support today (cross-project
   * dependency wiring belongs in the detail page after both tasks exist).
   * Empty string means "no project chosen" and the picker reads as
   * disabled chrome with a one-line nudge.
   */
  projectId: string;
  selected: string[];
  onChange: (next: string[]) => void;
  disabled: boolean;
};

export function TaskCreateDependsOnPicker({
  projectId,
  selected,
  onChange,
  disabled,
}: Props) {
  const inputId = useId();
  const listboxId = useId();
  const browseTitleId = useId();
  const picker = useTaskCreateDependsOnPickerState({
    projectId,
    selected,
    onChange,
    disabled,
  });

  return (
    <div className="task-create-scheduling__field task-create-deps">
      <FieldLabel htmlFor={inputId}>Depends on</FieldLabel>
      <TaskCreateDependsOnSearchRow
        inputId={inputId}
        listboxId={listboxId}
        hasProject={picker.hasProject}
        listOpen={picker.listOpen}
        query={picker.query}
        inputDisabled={picker.inputDisabled}
        projectTaskCount={picker.projectTasks.length}
        onQueryChange={picker.handleQueryChange}
        onFocus={picker.handleInputFocus}
        onBlur={picker.handleInputBlur}
        onKeyDown={picker.handleInputKeyDown}
        onBrowseOpen={() => picker.setBrowseOpen(true)}
      />

      {picker.listOpen && picker.hasProject ? (
        <TaskCreateDependsOnTypeaheadList
          listboxId={listboxId}
          typeaheadResults={picker.typeaheadResults}
          projectTaskCount={picker.projectTasks.length}
          onSelect={picker.handleSelectFromTypeahead}
        />
      ) : null}

      <TaskCreateDependsOnSelectedChips
        selected={selected}
        labelLookup={picker.labelLookup}
        disabled={disabled}
        onRemove={picker.removeId}
      />

      <p className="hint">{picker.helperCopy}</p>

      {picker.browseOpen ? (
        <TaskCreateDependsOnBrowseModal
          browseTitleId={browseTitleId}
          browseQuery={picker.browseQuery}
          browseResults={picker.browseResults}
          selectedSet={picker.selectedSet}
          selectedCount={selected.length}
          disabled={disabled}
          onBrowseQueryChange={picker.setBrowseQuery}
          onClose={() => picker.setBrowseOpen(false)}
          onToggle={picker.toggleId}
        />
      ) : null}
    </div>
  );
}
