package tools

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/gardo984/mcp-server/src/devops_server/internal/framework"
)

// ListSkillsInput is the (empty) input of the list_skills tool.
type ListSkillsInput struct{}

// SkillInfo describes one skill available in the framework.
type SkillInfo struct {
	Name        string `json:"name" jsonschema:"name to use in framework://skills/{name}"`
	Description string `json:"description,omitempty"`
}

// ListSkillsOutput is the structured result of the list_skills tool.
type ListSkillsOutput struct {
	Skills []SkillInfo `json:"skills"`
}

func registerListSkills(s *mcp.Server, root string) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "list_skills",
		Title:       "List Skills",
		Description: "List the skills available in the AI framework. Read one with the framework://skills/{name} resource.",
	}, listSkills(root))
}

// listSkills scans the framework on every call, so the result never goes stale.
func listSkills(root string) func(context.Context, *mcp.CallToolRequest, ListSkillsInput) (*mcp.CallToolResult, ListSkillsOutput, error) {
	return func(_ context.Context, _ *mcp.CallToolRequest, _ ListSkillsInput) (*mcp.CallToolResult, ListSkillsOutput, error) {
		skills, err := framework.ScanSkills(root)
		if err != nil {
			return nil, ListSkillsOutput{}, err
		}
		out := ListSkillsOutput{Skills: make([]SkillInfo, 0, len(skills))}
		for _, s := range skills {
			out.Skills = append(out.Skills, SkillInfo{Name: s.Name, Description: s.Description})
		}
		return nil, out, nil
	}
}
