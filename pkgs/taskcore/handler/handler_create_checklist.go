package handler

import (
	"fmt"
	"strings"

	checklistcontract "github.com/AlexsanderHamir/Hamix/pkgs/taskchecklist/contract"
	checkliststore "github.com/AlexsanderHamir/Hamix/pkgs/taskchecklist/store"
	"github.com/AlexsanderHamir/Hamix/pkgs/taskcore/contract"
	"github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
)

// parseCreateChecklistItems normalizes POST /tasks checklist_items: trims text,
// drops blanks, and requires at least one surviving criterion.
//
//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func parseCreateChecklistItems(items []contract.CreateChecklistItemInput) ([]contract.CreateChecklistItemInput, error) {
	var out []contract.CreateChecklistItemInput
	for _, it := range items {
		t := strings.TrimSpace(it.Text)
		if t == "" {
			continue
		}
		cmds := make([]checklistcontract.VerifyCommandInput, 0, len(it.VerifyCommands))
		for _, c := range it.VerifyCommands {
			cmds = append(cmds, checklistcontract.VerifyCommandInput{
				Command:         c.Command,
				ExpectedOutcome: c.ExpectedOutcome,
			})
		}
		normalized, err := checkliststore.NormalizeVerifyCommands(cmds)
		if err != nil {
			return nil, err
		}
		out = append(out, contract.CreateChecklistItemInput{
			Text:           t,
			VerifyCommands: normalized,
		})
	}
	if len(out) < 1 {
		return nil, fmt.Errorf("%w: at least one done criterion required", domain.ErrInvalidInput)
	}
	return out, nil
}
