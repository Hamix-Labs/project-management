package handler

import "strings"

// testCriterionText is the default non-empty done criterion used in handler
// contract tests after POST /tasks began requiring checklist_items.
const testCriterionText = "Test criterion"

// withCreateChecklist injects the required checklist_items field into a POST
// /tasks JSON object and, when a test git repo is seeded, project_id and
// worktree_id. jsonBody must be a single object literal ending with `}`.
// Deprecated: use withCreateChecklistForURL when baseURL is known.
func withCreateChecklist(jsonBody string) string {
	jsonBody = strings.TrimSpace(jsonBody)
	if strings.Contains(jsonBody, "checklist_items") {
		return jsonBody
	}
	if !strings.HasSuffix(jsonBody, "}") {
		return jsonBody
	}
	return jsonBody[:len(jsonBody)-1] + `,"checklist_items":[{"text":"` + testCriterionText + `"}]}`
}
