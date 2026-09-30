package mcpserver

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/gardo984/mcp-server/src/devops_server/internal/config"
)

// newTestClient starts the server over an in-memory transport and returns a
// connected client session.
func newTestClient(t *testing.T, root string) *mcp.ClientSession {
	t.Helper()
	srv, err := New(config.Config{FrameworkRoot: root}, slog.New(slog.NewTextHandler(os.Stderr, nil)))
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	t1, t2 := mcp.NewInMemoryTransports()
	if _, err := srv.mcp.Connect(ctx, t1, nil); err != nil {
		t.Fatal(err)
	}
	c := mcp.NewClient(&mcp.Implementation{Name: "test", Version: "v0"}, nil)
	cs, err := c.Connect(ctx, t2, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { cs.Close() })
	return cs
}

// writeAgent drops an agent file under the framework's .github/agents dir.
func writeAgent(t *testing.T, root, file, body string) {
	t.Helper()
	dir := filepath.Join(root, ".github", "agents")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, file), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

// listAgentsText calls the list_agents tool and returns its text content.
func listAgentsText(t *testing.T, cs *mcp.ClientSession) string {
	t.Helper()
	res, err := cs.CallTool(context.Background(), &mcp.CallToolParams{Name: "list_agents"})
	if err != nil {
		t.Fatal(err)
	}
	return res.Content[0].(*mcp.TextContent).Text
}

// writePrompt drops a prompt file under the framework's .github/prompts dir.
func writePrompt(t *testing.T, root, file, body string) {
	t.Helper()
	dir := filepath.Join(root, ".github", "prompts")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, file), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

// listPromptsText calls the list_prompts tool and returns its text content.
func listPromptsText(t *testing.T, cs *mcp.ClientSession) string {
	t.Helper()
	res, err := cs.CallTool(context.Background(), &mcp.CallToolParams{Name: "list_prompts"})
	if err != nil {
		t.Fatal(err)
	}
	return res.Content[0].(*mcp.TextContent).Text
}

// writeSkill drops a directory-per-skill under a hub's skills/ dir.
func writeSkill(t *testing.T, root, hub, name, body string) {
	t.Helper()
	dir := filepath.Join(root, hub, "skills", name)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

// listSkillsText calls the list_skills tool and returns its text content.
func listSkillsText(t *testing.T, cs *mcp.ClientSession) string {
	t.Helper()
	res, err := cs.CallTool(context.Background(), &mcp.CallToolParams{Name: "list_skills"})
	if err != nil {
		t.Fatal(err)
	}
	return res.Content[0].(*mcp.TextContent).Text
}

// listMemoriesText calls the list_memories tool and returns its text content.
func listMemoriesText(t *testing.T, cs *mcp.ClientSession) string {
	t.Helper()
	res, err := cs.CallTool(context.Background(), &mcp.CallToolParams{Name: "list_memories"})
	if err != nil {
		t.Fatal(err)
	}
	return res.Content[0].(*mcp.TextContent).Text
}

func TestSmoke(t *testing.T) {
	root := t.TempDir()
	writeAgent(t, root, "reviewer.agent.md",
		"---\nname: reviewer\ndescription: Reviews code changes.\n---\n\n# Reviewer agent\n")

	cs := newTestClient(t, root)
	ctx := context.Background()

	// No concrete resources are registered: agents are served through the
	// template, so nothing can go stale.
	for res, err := range cs.Resources(ctx, nil) {
		if err != nil {
			t.Fatal(err)
		}
		t.Errorf("unexpected concrete resource %s; agents should be template-only", res.URI)
	}

	// Every artifact kind is advertised as a URI template, and no concrete
	// resource is registered, so nothing can go stale.
	templates := map[string]bool{}
	for rt, err := range cs.ResourceTemplates(ctx, nil) {
		if err != nil {
			t.Fatal(err)
		}
		templates[rt.URITemplate] = true
	}
	for _, want := range []string{"framework://agents/{name}", "framework://prompts/{name}", "framework://skills/{name}", "framework://memories/{name}"} {
		if !templates[want] {
			t.Errorf("missing resource template %q (got %v)", want, templates)
		}
	}

	// list_agents discovers what names the template accepts.
	text := listAgentsText(t, cs)
	if !strings.Contains(text, "reviewer") {
		t.Errorf("list_agents output = %q, want it to mention reviewer", text)
	}
	t.Logf("LIST AGENTS: %s", text)

	// The agent body is read through the template.
	rr, err := cs.ReadResource(ctx, &mcp.ReadResourceParams{URI: "framework://agents/reviewer"})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(rr.Contents[0].Text, "# Reviewer agent") {
		t.Errorf("resource body = %q, want agent content", rr.Contents[0].Text)
	}
	t.Logf("RESOURCE BODY: %q", rr.Contents[0].Text)
}

// TestListAgentsWithoutRestart proves that agents added after startup are
// visible to both the list_agents tool and the resource template.
func TestListAgentsWithoutRestart(t *testing.T) {
	root := t.TempDir()
	writeAgent(t, root, "first.agent.md", "---\ndescription: first\n---\n# First\n")

	cs := newTestClient(t, root)
	if got := listAgentsText(t, cs); strings.Contains(got, "second") {
		t.Fatalf("second agent should not exist yet: %s", got)
	}

	writeAgent(t, root, "second.agent.md", "---\ndescription: second\n---\n# Second\n")

	if got := listAgentsText(t, cs); !strings.Contains(got, "second") {
		t.Errorf("list_agents after adding second = %q, want it to mention second", got)
	}
	rr, err := cs.ReadResource(context.Background(), &mcp.ReadResourceParams{URI: "framework://agents/second"})
	if err != nil {
		t.Fatalf("read freshly added agent: %v", err)
	}
	if !strings.Contains(rr.Contents[0].Text, "# Second") {
		t.Errorf("resource body = %q, want new agent content", rr.Contents[0].Text)
	}
}

// TestListPrompts checks that prompts are discovered and can be read through
// the framework://prompts/{name} template.
func TestListPrompts(t *testing.T) {
	root := t.TempDir()
	writePrompt(t, root, "commit-changes.prompt.md",
		"---\nname: \"commit-changes\"\ndescription: \"Stages and commits.\"\n---\n\n# Commit\n")

	cs := newTestClient(t, root)
	text := listPromptsText(t, cs)
	if !strings.Contains(text, "commit-changes") {
		t.Errorf("list_prompts output = %q, want it to mention commit-changes", text)
	}
	t.Logf("LIST PROMPTS: %s", text)

	rr, err := cs.ReadResource(context.Background(), &mcp.ReadResourceParams{URI: "framework://prompts/commit-changes"})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(rr.Contents[0].Text, "# Commit") {
		t.Errorf("prompt body = %q, want prompt content", rr.Contents[0].Text)
	}
	t.Logf("PROMPT BODY: %q", rr.Contents[0].Text)
}

// TestListSkills checks that skills stored as <hub>/skills/<name>/SKILL.md are
// discovered through the list_skills tool and read through the
// framework://skills/{name} template.
func TestListSkills(t *testing.T) {
	root := t.TempDir()
	writeSkill(t, root, "pi-hub", "push-changes",
		"---\nname: push-changes\ndescription: Publishes commits.\n---\n\n# Push\n")

	cs := newTestClient(t, root)
	text := listSkillsText(t, cs)
	if !strings.Contains(text, "push-changes") {
		t.Errorf("list_skills output = %q, want it to mention push-changes", text)
	}
	t.Logf("LIST SKILLS: %s", text)

	rr, err := cs.ReadResource(context.Background(), &mcp.ReadResourceParams{URI: "framework://skills/push-changes"})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(rr.Contents[0].Text, "# Push") {
		t.Errorf("skill body = %q, want skill content", rr.Contents[0].Text)
	}
	t.Logf("SKILL BODY: %q", rr.Contents[0].Text)
}

// TestListMemories checks that memories are discovered through list_memories
// and read through framework://memories/{name}, and that a checkout without a
// memories directory simply lists none.
func TestListMemories(t *testing.T) {
	root := t.TempDir()

	cs := newTestClient(t, root)
	if got := listMemoriesText(t, cs); !strings.Contains(got, "\"memories\":[]") {
		t.Errorf("list_memories on empty root = %q, want empty list", got)
	}

	dir := filepath.Join(root, "memories")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "conventions.md"),
		[]byte("---\nname: conventions\ndescription: House conventions.\n---\n\n# Conventions\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	text := listMemoriesText(t, cs)
	if !strings.Contains(text, "conventions") {
		t.Errorf("list_memories output = %q, want it to mention conventions", text)
	}
	t.Logf("LIST MEMORIES: %s", text)

	rr, err := cs.ReadResource(context.Background(), &mcp.ReadResourceParams{URI: "framework://memories/conventions"})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(rr.Contents[0].Text, "# Conventions") {
		t.Errorf("memory body = %q, want memory content", rr.Contents[0].Text)
	}
	t.Logf("MEMORY BODY: %q", rr.Contents[0].Text)
}
