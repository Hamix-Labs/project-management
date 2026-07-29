package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/AlexsanderHamir/Hamix/pkgs/agents/agentmcp"
)

func main() {
	bindPath := flag.String("bind", "", "path to agent-tool-bind.json")
	flag.Parse()
	if strings.TrimSpace(*bindPath) == "" {
		fmt.Fprintln(os.Stderr, "hamix-agent-mcp: --bind is required")
		os.Exit(2)
	}
	sess, err := agentmcp.LoadBind(*bindPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "hamix-agent-mcp: load bind: %v\n", err)
		os.Exit(1)
	}
	if err := agentmcp.RunStdio(context.Background(), sess); err != nil {
		fmt.Fprintf(os.Stderr, "hamix-agent-mcp: %v\n", err)
		os.Exit(1)
	}
}
