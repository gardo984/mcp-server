package tools

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/gardo984/mcp-server/src/devops_server/internal/framework"
)

// ListAgentsInput is the (empty) input of the list_agents tool.
type ListAgentsInput struct{}

// AgentInfo describes one agent available in the framework.
type AgentInfo struct {
	Name        string `json:"name" jsonschema:"name to use in framework://agents/{name}"`
	Description string `json:"description,omitempty"`
}

// ListAgentsOutput is the structured result of the list_agents tool.
type ListAgentsOutput struct {
	Agents []AgentInfo `json:"agents"`
}

func registerListAgents(s *mcp.Server, root string) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "list_agents",
		Title:       "List Agents",
		Description: "List the agents available in the AI framework. Read one with the framework://agents/{name} resource.",
	}, listAgents(root))
}

// listAgents scans the framework on every call, so the result never goes stale.
func listAgents(root string) func(context.Context, *mcp.CallToolRequest, ListAgentsInput) (*mcp.CallToolResult, ListAgentsOutput, error) {
	return func(_ context.Context, _ *mcp.CallToolRequest, _ ListAgentsInput) (*mcp.CallToolResult, ListAgentsOutput, error) {
		agents, err := framework.ScanAgents(root)
		if err != nil {
			return nil, ListAgentsOutput{}, err
		}
		out := ListAgentsOutput{Agents: make([]AgentInfo, 0, len(agents))}
		for _, a := range agents {
			out.Agents = append(out.Agents, AgentInfo{Name: a.Name, Description: a.Description})
		}
		return nil, out, nil
	}
}
