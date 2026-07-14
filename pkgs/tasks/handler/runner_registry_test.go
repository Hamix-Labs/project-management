package handler

import (
	// Register production runners so handler whitebox HTTP/SSE tests that POST
	// /tasks can resolve settings.Runner after runners_contract_test moved out.
	_ "github.com/AlexsanderHamir/Hamix/pkgs/agents/runner/registry/all"
)
