package agentmcp

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Tool is one Hamix MCP capability. Register new tools via DefaultTools.
type Tool interface {
	Name() string
	Group() string
	Description() string
	Register(server *mcp.Server, sess *Session)
}

// Registry holds tools for one server build.
type Registry struct {
	tools []Tool
}

// NewRegistry returns an empty registry.
func NewRegistry() *Registry {
	return &Registry{}
}

// Add appends tools.
func (r *Registry) Add(tools ...Tool) {
	r.tools = append(r.tools, tools...)
}

// RegisterAll registers every tool on the MCP server.
func (r *Registry) RegisterAll(server *mcp.Server, sess *Session) {
	for _, t := range r.tools {
		t.Register(server, sess)
	}
}

// DefaultTools returns the v1 report submit tools and commit register tool.
func DefaultTools() []Tool {
	return []Tool{
		commitTool{},
		createPullRequestTool{},
		submitCriteriaTool{},
	}
}

// NewServer builds an MCP server with DefaultTools bound to sess.
func NewServer(sess *Session) *mcp.Server {
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "hamix-agent-mcp",
		Version: "v1.0.0",
	}, nil)
	reg := NewRegistry()
	reg.Add(DefaultTools()...)
	reg.RegisterAll(server, sess)
	return server
}

// RunStdio serves MCP over stdin/stdout until the client disconnects.
func RunStdio(ctx context.Context, sess *Session) error {
	server := NewServer(sess)
	return server.Run(ctx, &mcp.StdioTransport{})
}
