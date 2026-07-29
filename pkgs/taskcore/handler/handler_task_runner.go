package handler

import (
	"fmt"
	"strings"

	settingsdomain "github.com/AlexsanderHamir/Hamix/pkgs/settings/domain"
	"github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
)

// resolveTaskRunnerModel merges optional JSON fields with app settings and
// validates the runner id via the injected RunnerValidator.
//
//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func resolveTaskRunnerModel(body *taskCreateJSON, settings settingsdomain.AppSettings, runners RunnerValidator) (runner, cursorModel string, err error) {
	return resolveRunnerModelFields(body.Runner, body.CursorModel, settings, runners)
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func resolveRunnerModelFields(runnerPtr, cursorModelPtr *string, settings settingsdomain.AppSettings, runners RunnerValidator) (runner, cursorModel string, err error) {
	if runnerPtr != nil && strings.TrimSpace(*runnerPtr) != "" {
		runner = strings.TrimSpace(*runnerPtr)
	} else {
		runner = strings.TrimSpace(settings.Runner)
	}
	if runner == "" {
		runner = settingsdomain.DefaultRunner
	}
	if runners != nil {
		if lerr := runners.ValidateRunner(runner); lerr != nil {
			return "", "", fmt.Errorf("%w: runner", domain.ErrInvalidInput)
		}
	}
	if cursorModelPtr != nil {
		cursorModel = strings.TrimSpace(*cursorModelPtr)
	} else {
		cursorModel = strings.TrimSpace(settings.CursorModel)
	}
	return runner, cursorModel, nil
}
