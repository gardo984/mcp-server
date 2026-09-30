package tools

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/gardo984/mcp-server/src/devops_server/internal/framework"
)

// ListPromptsInput is the (empty) input of the list_prompts tool.
type ListPromptsInput struct{}

// PromptInfo describes one prompt available in the framework.
type PromptInfo struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

// ListPromptsOutput is the structured result of the list_prompts tool.
type ListPromptsOutput struct {
	Prompts []PromptInfo `json:"prompts"`
}

func registerListPrompts(s *mcp.Server, root string) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "list_prompts",
		Title:       "List Prompts",
		Description: "List the prompts available in the AI framework repository.",
	}, listPrompts(root))
}

// listPrompts scans the framework on every call, so the result never goes stale.
func listPrompts(root string) func(context.Context, *mcp.CallToolRequest, ListPromptsInput) (*mcp.CallToolResult, ListPromptsOutput, error) {
	return func(_ context.Context, _ *mcp.CallToolRequest, _ ListPromptsInput) (*mcp.CallToolResult, ListPromptsOutput, error) {
		prompts, err := framework.ScanPrompts(root)
		if err != nil {
			return nil, ListPromptsOutput{}, err
		}
		out := ListPromptsOutput{Prompts: make([]PromptInfo, 0, len(prompts))}
		for _, p := range prompts {
			out.Prompts = append(out.Prompts, PromptInfo{Name: p.Name, Description: p.Description})
		}
		return nil, out, nil
	}
}
