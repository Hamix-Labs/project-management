package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"

	draftmcp "github.com/AlexsanderHamir/Hamix/pkgs/draftassist/mcp"
	draftassiststore "github.com/AlexsanderHamir/Hamix/pkgs/draftassist/store"
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
	// In-process store for local smoke; production Plan 3 will call taskapi HTTP.
	// Tools fail closed on nonce mismatch against the bind nonce when the
	// session exists in this process (dev). For a cold binary with no prior
	// session, CreateSession is not automatic — operators use taskapi HTTP.
	store := draftassiststore.NewMemoryStore()
	host := &draftmcp.ToolHost{Bind: bind, Store: store}
	if err := draftmcp.RunStdio(context.Background(), host); err != nil {
		fmt.Fprintf(os.Stderr, "hamix-draft-mcp: %v\n", err)
		os.Exit(1)
	}
}
