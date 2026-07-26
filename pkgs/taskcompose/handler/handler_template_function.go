package handler

import (
	"encoding/json"
	"fmt"
	"strings"

	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	taskcorehandler "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/handler"
)

const maxFunctionBindingValues = 10

type functionBindingJSON struct {
	InputID   string                   `json:"input_id"`
	Paths     []string                 `json:"paths,omitempty"`
	Functions []functionRefBindingJSON `json:"functions,omitempty"`
}

type functionRefBindingJSON struct {
	Path string `json:"path"`
	Name string `json:"name"`
	Line int    `json:"line"`
}

// applyFunctionBindingsToPayload validates bindings against the schema, appends a
// soft-scope block to initial_prompt, and clears function_inputs for task create.
func applyFunctionBindingsToPayload(raw json.RawMessage, bindings []functionBindingJSON) (json.RawMessage, error) {
	payload, err := taskcorehandler.DecodeComposePayload(raw)
	if err != nil {
		return nil, err
	}
	schema := payload.FunctionInputs
	if len(schema) == 0 {
		if len(bindings) > 0 {
			return nil, fmt.Errorf("%w: function_bindings not allowed without function_inputs", taskcoredomain.ErrInvalidInput)
		}
		return raw, nil
	}
	byID := make(map[string]functionBindingJSON, len(bindings))
	for _, b := range bindings {
		id := strings.TrimSpace(b.InputID)
		if id == "" {
			return nil, fmt.Errorf("%w: function_bindings.input_id required", taskcoredomain.ErrInvalidInput)
		}
		if _, dup := byID[id]; dup {
			return nil, fmt.Errorf("%w: duplicate function_bindings input_id %q", taskcoredomain.ErrInvalidInput, id)
		}
		byID[id] = b
	}
	for id := range byID {
		found := false
		for _, def := range schema {
			if strings.TrimSpace(def.ID) == id {
				found = true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("%w: unknown function_bindings input_id %q", taskcoredomain.ErrInvalidInput, id)
		}
	}

	var dirs, files []string
	var funcs []functionRefBindingJSON
	for _, def := range schema {
		id := strings.TrimSpace(def.ID)
		b, ok := byID[id]
		required := def.FunctionInputRequired()
		if !ok {
			if required {
				return nil, fmt.Errorf("%w: missing binding for function input %q", taskcoredomain.ErrInvalidInput, id)
			}
			continue
		}
		switch def.Kind {
		case taskcorehandler.FunctionInputKindDir:
			paths, err := normalizeBindingPaths(b.Paths, def.Multiple, "dir")
			if err != nil {
				return nil, fmt.Errorf("%w: input %q: %s", taskcoredomain.ErrInvalidInput, id, err.Error())
			}
			if len(b.Functions) > 0 {
				return nil, fmt.Errorf("%w: input %q: functions not allowed for dir", taskcoredomain.ErrInvalidInput, id)
			}
			dirs = append(dirs, paths...)
		case taskcorehandler.FunctionInputKindFile:
			paths, err := normalizeBindingPaths(b.Paths, def.Multiple, "file")
			if err != nil {
				return nil, fmt.Errorf("%w: input %q: %s", taskcoredomain.ErrInvalidInput, id, err.Error())
			}
			if len(b.Functions) > 0 {
				return nil, fmt.Errorf("%w: input %q: functions not allowed for file", taskcoredomain.ErrInvalidInput, id)
			}
			files = append(files, paths...)
		case taskcorehandler.FunctionInputKindFunction:
			refs, err := normalizeBindingFunctions(b.Functions, def.Multiple)
			if err != nil {
				return nil, fmt.Errorf("%w: input %q: %s", taskcoredomain.ErrInvalidInput, id, err.Error())
			}
			if len(b.Paths) > 0 {
				return nil, fmt.Errorf("%w: input %q: paths not allowed for function", taskcoredomain.ErrInvalidInput, id)
			}
			funcs = append(funcs, refs...)
		default:
			return nil, fmt.Errorf("%w: input %q: invalid kind", taskcoredomain.ErrInvalidInput, id)
		}
	}

	payload.InitialPrompt = appendSoftScope(payload.InitialPrompt, dirs, files, funcs)
	payload.FunctionInputs = nil
	return taskcorehandler.ComposePayloadToRaw(payload)
}

func normalizeBindingPaths(paths []string, multiple bool, kind string) ([]string, error) {
	out := make([]string, 0, len(paths))
	seen := make(map[string]struct{}, len(paths))
	for _, p := range paths {
		p = strings.TrimSpace(strings.ReplaceAll(p, "\\", "/"))
		p = strings.TrimPrefix(p, "/")
		if p == "" {
			continue
		}
		if _, ok := seen[p]; ok {
			continue
		}
		seen[p] = struct{}{}
		out = append(out, p)
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("%s path required", kind)
	}
	if !multiple && len(out) > 1 {
		return nil, fmt.Errorf("%s input does not allow multiple values", kind)
	}
	if len(out) > maxFunctionBindingValues {
		return nil, fmt.Errorf("at most %d %s values", maxFunctionBindingValues, kind)
	}
	return out, nil
}

func normalizeBindingFunctions(refs []functionRefBindingJSON, multiple bool) ([]functionRefBindingJSON, error) {
	out := make([]functionRefBindingJSON, 0, len(refs))
	for _, r := range refs {
		path := strings.TrimSpace(strings.ReplaceAll(r.Path, "\\", "/"))
		path = strings.TrimPrefix(path, "/")
		name := strings.TrimSpace(r.Name)
		if path == "" || name == "" || r.Line < 1 {
			return nil, fmt.Errorf("function requires path, name, and line >= 1")
		}
		out = append(out, functionRefBindingJSON{Path: path, Name: name, Line: r.Line})
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("function reference required")
	}
	if !multiple && len(out) > 1 {
		return nil, fmt.Errorf("function input does not allow multiple values")
	}
	if len(out) > maxFunctionBindingValues {
		return nil, fmt.Errorf("at most %d function values", maxFunctionBindingValues)
	}
	return out, nil
}

func appendSoftScope(prompt string, dirs, files []string, funcs []functionRefBindingJSON) string {
	var b strings.Builder
	b.WriteString(strings.TrimRight(prompt, "\n"))
	if b.Len() > 0 {
		b.WriteString("\n\n")
	}
	b.WriteString("## Scope (do not expand beyond)\n")
	if len(dirs) > 0 {
		b.WriteString("- Directories:")
		for _, d := range dirs {
			b.WriteString(" `")
			b.WriteString(d)
			b.WriteString("`")
		}
		b.WriteByte('\n')
	}
	if len(files) > 0 {
		b.WriteString("- Files:")
		for _, f := range files {
			b.WriteString(" @")
			b.WriteString(f)
		}
		b.WriteByte('\n')
	}
	if len(funcs) > 0 {
		b.WriteString("- Functions:")
		for _, fn := range funcs {
			fmt.Fprintf(&b, " @%s(%d-%d) (`%s`)", fn.Path, fn.Line, fn.Line, fn.Name)
		}
		b.WriteByte('\n')
	}
	b.WriteString("Restrict all edits and investigation to the paths above unless the operator expands scope.\n")
	return b.String()
}
