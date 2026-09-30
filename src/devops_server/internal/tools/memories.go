package tools

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/gardo984/mcp-server/src/devops_server/internal/framework"
)

// ListMemoriesInput is the (empty) input of the list_memories tool.
type ListMemoriesInput struct{}

// MemoryInfo describes one memory available in the framework.
type MemoryInfo struct {
	Name        string `json:"name" jsonschema:"name to use in framework://memories/{name}"`
	Description string `json:"description,omitempty"`
}

// ListMemoriesOutput is the structured result of the list_memories tool.
type ListMemoriesOutput struct {
	Memories []MemoryInfo `json:"memories"`
}

func registerListMemories(s *mcp.Server, root string) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "list_memories",
		Title:       "List Memories",
		Description: "List the memories available in the AI framework. Read one with the framework://memories/{name} resource.",
	}, listMemories(root))
}

// listMemories scans the framework on every call, so the result never goes stale.
func listMemories(root string) func(context.Context, *mcp.CallToolRequest, ListMemoriesInput) (*mcp.CallToolResult, ListMemoriesOutput, error) {
	return func(_ context.Context, _ *mcp.CallToolRequest, _ ListMemoriesInput) (*mcp.CallToolResult, ListMemoriesOutput, error) {
		memories, err := framework.ScanMemories(root)
		if err != nil {
			return nil, ListMemoriesOutput{}, err
		}
		out := ListMemoriesOutput{Memories: make([]MemoryInfo, 0, len(memories))}
		for _, m := range memories {
			out.Memories = append(out.Memories, MemoryInfo{Name: m.Name, Description: m.Description})
		}
		return nil, out, nil
	}
}
