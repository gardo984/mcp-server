// Package tools registers the MCP tools exposed by the server. They exist so a
// client can discover the names accepted by the framework resource templates.
package tools

import (
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Register adds every framework tool to the server.
func Register(s *mcp.Server, root string) {
	registerListAgents(s, root)
	registerListPrompts(s, root)
	registerListSkills(s, root)
	registerListMemories(s, root)
}
