package handler

import (
	"encoding/json"
	"fmt"
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	"net/url"
	"strings"
)

type taskDraftSaveJSON struct {
	ID      string          `json:"id"`
	Name    string          `json:"name"`
	Payload json.RawMessage `json:"payload"`
}

type taskTemplateSaveJSON struct {
	ID      string          `json:"id"`
	Name    string          `json:"name"`
	Payload json.RawMessage `json:"payload"`
}

type taskTemplatePatchJSON struct {
	Name    *string         `json:"name"`
	Payload json.RawMessage `json:"payload"`
}

type taskTemplateInstantiateItemJSON struct {
	TemplateID       string                `json:"template_id"`
	Count            *int                  `json:"count,omitempty"`
	FunctionBindings []functionBindingJSON `json:"function_bindings,omitempty"`
}

type taskTemplateInstantiateJSON struct {
	TemplateIDs []string                          `json:"template_ids,omitempty"`
	Count       *int                              `json:"count,omitempty"`
	Items       []taskTemplateInstantiateItemJSON `json:"items,omitempty"`
}

type taskTemplateInstantiateItem struct {
	TemplateID       string
	Count            int
	FunctionBindings []functionBindingJSON
}

type taskTemplateInstantiateErrorJSON struct {
	TemplateID string `json:"template_id"`
	Error      string `json:"error"`
}

type taskTemplateInstantiateAcceptedJSON struct {
	Accepted bool                               `json:"accepted"`
	Total    int                                `json:"total"`
	Errors   []taskTemplateInstantiateErrorJSON `json:"errors"`
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by listTaskTemplates."
func parseTemplateListQuery(q url.Values) (sort, order, tag string, err error) {
	sort = strings.TrimSpace(q.Get("sort"))
	if sort == "" {
		sort = "updated_at"
	} else {
		switch sort {
		case "updated_at", "name", "instantiate_count":
		default:
			return "", "", "", fmt.Errorf("%w: invalid sort %q", taskcoredomain.ErrInvalidInput, sort)
		}
	}
	order = strings.ToLower(strings.TrimSpace(q.Get("order")))
	if order == "" {
		order = "desc"
	} else {
		switch order {
		case "asc", "desc":
		default:
			return "", "", "", fmt.Errorf("%w: invalid order %q", taskcoredomain.ErrInvalidInput, order)
		}
	}
	tag = strings.TrimSpace(q.Get("tag"))
	return sort, strings.ToUpper(order), tag, nil
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func resolveInstantiateCount(raw *int) (int, error) {
	if raw == nil {
		return 1, nil
	}
	if *raw < 1 || *raw > maxTemplateInstantiateCountPerItem {
		return 0, fmt.Errorf("%w: count must be integer 1..%d", taskcoredomain.ErrInvalidInput, maxTemplateInstantiateCountPerItem)
	}
	return *raw, nil
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func normalizeInstantiateItems(body taskTemplateInstantiateJSON) ([]taskTemplateInstantiateItem, error) {
	if len(body.Items) > 0 {
		items := make([]taskTemplateInstantiateItem, 0, len(body.Items))
		seen := make(map[string]struct{}, len(body.Items))
		total := 0
		for _, row := range body.Items {
			templateID := strings.TrimSpace(row.TemplateID)
			if templateID == "" {
				return nil, fmt.Errorf("%w: template id required", taskcoredomain.ErrInvalidInput)
			}
			if _, dup := seen[templateID]; dup {
				return nil, fmt.Errorf("%w: duplicate template_id %q in items", taskcoredomain.ErrInvalidInput, templateID)
			}
			seen[templateID] = struct{}{}
			count, err := resolveInstantiateCount(row.Count)
			if err != nil {
				return nil, err
			}
			total += count
			if total > maxTemplateInstantiateTotalCreates {
				return nil, fmt.Errorf("%w: total creates must not exceed %d", taskcoredomain.ErrInvalidInput, maxTemplateInstantiateTotalCreates)
			}
			items = append(items, taskTemplateInstantiateItem{
				TemplateID:       templateID,
				Count:            count,
				FunctionBindings: row.FunctionBindings,
			})
		}
		if len(items) == 0 {
			return nil, fmt.Errorf("%w: items required", taskcoredomain.ErrInvalidInput)
		}
		return items, nil
	}

	if len(body.TemplateIDs) == 0 {
		return nil, fmt.Errorf("%w: template_ids or items required", taskcoredomain.ErrInvalidInput)
	}
	defaultCount, err := resolveInstantiateCount(body.Count)
	if err != nil {
		return nil, err
	}
	total := defaultCount * len(body.TemplateIDs)
	if total > maxTemplateInstantiateTotalCreates {
		return nil, fmt.Errorf("%w: total creates must not exceed %d", taskcoredomain.ErrInvalidInput, maxTemplateInstantiateTotalCreates)
	}
	items := make([]taskTemplateInstantiateItem, 0, len(body.TemplateIDs))
	for _, templateID := range body.TemplateIDs {
		templateID = strings.TrimSpace(templateID)
		if templateID == "" {
			return nil, fmt.Errorf("%w: template id required", taskcoredomain.ErrInvalidInput)
		}
		items = append(items, taskTemplateInstantiateItem{
			TemplateID: templateID,
			Count:      defaultCount,
		})
	}
	return items, nil
}
