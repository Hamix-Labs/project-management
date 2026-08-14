package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"

	draftmcp "github.com/AlexsanderHamir/Hamix/pkgs/draftassist/mcp"
)

//funclogmeasure:skip category=tool-required-noop reason="cmd entrypoint; slog JSON sink is configured elsewhere."
func main() {
	bindPath := flag.String("bind", "", "path to draft-assist bind JSON")
	flag.Parse()
	if strings.TrimSpace(*bindPath) == "" {
		fmt.Fprintln(os.Stderr, "hamix-draft-mcp: --bind is required")
		os.Exit(2)
	}
	bind, err := draftmcp.LoadBind(*bindPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "hamix-draft-mcp: load bind: %v\n", err)
		os.Exit(1)
	}
	if strings.TrimSpace(bind.TaskAPIBaseURL) == "" {
		fmt.Fprintln(os.Stderr, "hamix-draft-mcp: bind.taskapi_base_url is required (Plan 4 dropped the in-process store)")
		os.Exit(1)
	}
	client := draftmcp.New(bind.TaskAPIBaseURL, bind.SessionID, bind.Nonce, nil)
	host := &draftmcp.ToolHost{Bind: bind, Client: client}
	if err := draftmcp.RunStdio(context.Background(), host); err != nil {
		fmt.Fprintf(os.Stderr, "hamix-draft-mcp: %v\n", err)
		os.Exit(1)
	}
}
