package handler

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
)

const maxFunctionInputDefs = 20

var functionInputIDPattern = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)

// FunctionInputKind is a template function input slot kind.
type FunctionInputKind string

const (
	FunctionInputKindDir      FunctionInputKind = "dir"
	FunctionInputKindFile     FunctionInputKind = "file"
	FunctionInputKindFunction FunctionInputKind = "function"
)

// FunctionInputDef is one create-time input slot declared on a template.
type FunctionInputDef struct {
	ID       string            `json:"id"`
	Kind     FunctionInputKind `json:"kind"`
	Label    string            `json:"label"`
	Required *bool             `json:"required,omitempty"`
	Multiple bool              `json:"multiple,omitempty"`
}

// FunctionInputRequired reports whether the slot must be bound (default true).
func (d FunctionInputDef) FunctionInputRequired() bool {
	if d.Required == nil {
		return true
	}
	return *d.Required
}

// ValidateFunctionInputsSchema validates template function_inputs declarations.
//
//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by ValidateCompose."
func ValidateFunctionInputsSchema(inputs []FunctionInputDef) error {
	if len(inputs) == 0 {
		return nil
	}
	if len(inputs) > maxFunctionInputDefs {
		return fmt.Errorf("%w: function_inputs must have at most %d entries", domain.ErrInvalidInput, maxFunctionInputDefs)
	}
	seen := make(map[string]struct{}, len(inputs))
	for i, in := range inputs {
		id := strings.TrimSpace(in.ID)
		if id == "" {
			return fmt.Errorf("%w: function_inputs[%d].id required", domain.ErrInvalidInput, i)
		}
		if !functionInputIDPattern.MatchString(id) {
			return fmt.Errorf("%w: function_inputs[%d].id must match [a-z][a-z0-9_]*", domain.ErrInvalidInput, i)
		}
		if _, dup := seen[id]; dup {
			return fmt.Errorf("%w: duplicate function_inputs id %q", domain.ErrInvalidInput, id)
		}
		seen[id] = struct{}{}
		label := strings.TrimSpace(in.Label)
		if label == "" {
			return fmt.Errorf("%w: function_inputs[%d].label required", domain.ErrInvalidInput, i)
		}
		switch in.Kind {
		case FunctionInputKindDir, FunctionInputKindFile, FunctionInputKindFunction:
		default:
			return fmt.Errorf("%w: function_inputs[%d].kind must be dir, file, or function", domain.ErrInvalidInput, i)
		}
		if in.Kind == FunctionInputKindFunction && in.Multiple {
			return fmt.Errorf("%w: function_inputs[%d] kind function cannot be multiple", domain.ErrInvalidInput, i)
		}
	}
	return nil
}
