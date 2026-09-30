// Package resources exposes the AI framework repository through MCP resources,
// so any agent can pull framework artifacts on demand.
//
// Each artifact kind is exposed through a URI template, framework://<kind>/{name}.
// Nothing is registered per artifact at startup: every read hits the repository,
// so both the content and the set of available names stay fresh without a
// restart. Use the list_agents, list_prompts, list_skills and list_memories
// tools to discover names.
package resources

import (
	"context"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/gardo984/mcp-server/src/devops_server/internal/framework"
)

// Scheme is the URI scheme used for framework resources.
const Scheme = "framework"

const (
	agentsSegment   = "agents"
	promptsSegment  = "prompts"
	skillsSegment   = "skills"
	memoriesSegment = "memories"
)

// reader reads one framework artifact by name.
type reader func(root, name string) (string, error)

// kind couples a URI segment with the loader for that artifact kind.
type kind struct {
	segment     string
	name        string
	title       string
	description string
	read        reader
}

var kinds = []kind{
	{
		segment:     agentsSegment,
		name:        "ai-framework-agents",
		title:       "AI framework agent",
		description: "Read an agent from the AI framework repository. Use the list_agents tool to discover names.",
		read:        framework.LoadAgent,
	},
	{
		segment:     promptsSegment,
		name:        "ai-framework-prompts",
		title:       "AI framework prompt",
		description: "Read a prompt from the AI framework repository. Use the list_prompts tool to discover names.",
		read:        framework.LoadPrompt,
	},
	{
		segment:     skillsSegment,
		name:        "ai-framework-skills",
		title:       "AI framework skill",
		description: "Read a skill from the AI framework repository. Use the list_skills tool to discover names.",
		read:        framework.LoadSkill,
	},
	{
		segment:     memoriesSegment,
		name:        "ai-framework-memories",
		title:       "AI framework memory",
		description: "Read a memory from the AI framework repository. Use the list_memories tool to discover names.",
		read:        framework.LoadMemory,
	},
}

// Register wires the framework artifacts rooted at root into the MCP server, one
// URI template per kind. Clients read an artifact with framework://<kind>/{name}.
func Register(s *mcp.Server, root string) {
	for _, k := range kinds {
		s.AddResourceTemplate(&mcp.ResourceTemplate{
			URITemplate: Scheme + "://" + k.segment + "/{name}",
			Name:        k.name,
			Title:       k.title,
			Description: k.description,
			MIMEType:    "text/markdown",
		}, readHandler(root, k))
	}
}

func readHandler(root string, k kind) mcp.ResourceHandler {
	return func(_ context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
		name, err := parseURI(req.Params.URI, k.segment)
		if err != nil {
			return nil, err
		}
		body, err := k.read(root, name)
		if err != nil {
			return nil, err
		}
		return &mcp.ReadResourceResult{
			Contents: []*mcp.ResourceContents{{
				URI:      req.Params.URI,
				MIMEType: "text/markdown",
				Text:     body,
			}},
		}, nil
	}
}

// parseURI extracts the artifact name from framework://<segment>/{name}.
func parseURI(uri, segment string) (string, error) {
	prefix := Scheme + "://" + segment + "/"
	name, ok := strings.CutPrefix(uri, prefix)
	if !ok || name == "" {
		return "", fmt.Errorf("unsupported resource URI %q: want %s{name}", uri, prefix)
	}
	return name, nil
}
