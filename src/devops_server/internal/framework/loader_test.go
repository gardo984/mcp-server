package framework

import (
	"os"
	"path/filepath"
	"testing"
)

func write(t *testing.T, path, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestScanAgentsPrefersGitHub(t *testing.T) {
	root := t.TempDir()
	write(t, filepath.Join(root, ".github", "agents", "backend-sre.agent.md"),
		"---\nname: backend-sre\ndescription: 'Backend & SRE expert.'\n---\n\n# Backend\n")
	write(t, filepath.Join(root, "agents", "legacy.md"), "# Legacy\n")

	agents, err := ScanAgents(root)
	if err != nil {
		t.Fatal(err)
	}
	byName := map[string]Artifact{}
	for _, a := range agents {
		byName[a.Name] = a
	}

	got, ok := byName["backend-sre"]
	if !ok {
		t.Fatalf("agent backend-sre not found; got %v", byName)
	}
	if got.Description != "Backend & SRE expert." {
		t.Errorf("description = %q", got.Description)
	}
	if got.Path != filepath.Join(".github", "agents", "backend-sre.agent.md") {
		t.Errorf("path = %q", got.Path)
	}
	if _, ok := byName["legacy"]; !ok {
		t.Errorf("flat agents/ layout must still be scanned; got %v", byName)
	}
}

func TestScanDoesNotShadowAcrossDirs(t *testing.T) {
	root := t.TempDir()
	write(t, filepath.Join(root, ".github", "agents", "dup.agent.md"), "---\ndescription: high priority\n---\n")
	write(t, filepath.Join(root, "agents", "dup.md"), "---\ndescription: low priority\n---\n")

	agents, err := ScanAgents(root)
	if err != nil {
		t.Fatal(err)
	}
	var dup []Artifact
	for _, a := range agents {
		if a.Name == "dup" {
			dup = append(dup, a)
		}
	}
	if len(dup) != 1 {
		t.Fatalf("want exactly one dup agent, got %d", len(dup))
	}
	if dup[0].Description != "high priority" {
		t.Errorf("description = %q, want the .github one", dup[0].Description)
	}
}

func TestScanPrompts(t *testing.T) {
	root := t.TempDir()
	write(t, filepath.Join(root, ".github", "prompts", "commit-changes.prompt.md"),
		"---\nname: \"commit-changes\"\ndescription: \"Stages and commits.\"\n---\n\n# Commit\n")
	write(t, filepath.Join(root, "prompts", "legacy.md"), "# Legacy prompt\n")

	prompts, err := ScanPrompts(root)
	if err != nil {
		t.Fatal(err)
	}
	byName := map[string]Artifact{}
	for _, p := range prompts {
		byName[p.Name] = p
	}

	got, ok := byName["commit-changes"]
	if !ok {
		t.Fatalf("prompt commit-changes not found; got %v", byName)
	}
	if got.Description != "Stages and commits." {
		t.Errorf("description = %q", got.Description)
	}
	if got.Path != filepath.Join(".github", "prompts", "commit-changes.prompt.md") {
		t.Errorf("path = %q", got.Path)
	}
	if _, ok := byName["legacy"]; !ok {
		t.Errorf("flat prompts/ layout must still be scanned; got %v", byName)
	}
}

func TestLoadAgent(t *testing.T) {
	root := t.TempDir()
	body := "---\nname: router\ndescription: 'Routes work.'\n---\n\n# Router\n"
	write(t, filepath.Join(root, ".github", "agents", "router.agent.md"), body)

	got, err := LoadAgent(root, "router")
	if err != nil {
		t.Fatal(err)
	}
	if got != body {
		t.Errorf("LoadAgent = %q, want full body", got)
	}

	if _, err := LoadAgent(root, "../secrets"); err == nil {
		t.Error("path traversal must be rejected")
	}
	if _, err := LoadAgent(root, "missing"); err == nil {
		t.Error("missing agent must be an error")
	}
}

func TestLoadPrompt(t *testing.T) {
	root := t.TempDir()
	body := "---\nname: commit-changes\ndescription: 'Stages and commits.'\n---\n\n# Commit\n"
	write(t, filepath.Join(root, ".github", "prompts", "commit-changes.prompt.md"), body)

	got, err := LoadPrompt(root, "commit-changes")
	if err != nil {
		t.Fatal(err)
	}
	if got != body {
		t.Errorf("LoadPrompt = %q, want full body", got)
	}

	if _, err := LoadPrompt(root, "../secrets"); err == nil {
		t.Error("path traversal must be rejected")
	}
}

func TestScanSkills(t *testing.T) {
	root := t.TempDir()
	// Directory-per-skill layout under a hub, found through the "*/skills" glob.
	write(t, filepath.Join(root, "pi-hub", "skills", "push-changes", "SKILL.md"),
		"---\nname: push-changes\ndescription: 'Publishes commits.'\n---\n\n# Push\n")
	// Flat skill file under .github/skills, which must win by priority.
	write(t, filepath.Join(root, ".github", "skills", "push-changes.prompt.md"),
		"---\ndescription: high priority\n---\n# Push github\n")
	write(t, filepath.Join(root, "openclaw-hub", "skills", "refresh-repo", "SKILL.md"),
		"---\ndescription: Refreshes the repo.\n---\n# Refresh\n")

	skills, err := ScanSkills(root)
	if err != nil {
		t.Fatal(err)
	}
	byName := map[string]Artifact{}
	for _, s := range skills {
		byName[s.Name] = s
	}

	got, ok := byName["push-changes"]
	if !ok {
		t.Fatalf("skill push-changes not found; got %v", byName)
	}
	if got.Description != "high priority" {
		t.Errorf("description = %q, want the .github one to win", got.Description)
	}
	if got.Path != filepath.Join(".github", "skills", "push-changes.prompt.md") {
		t.Errorf("path = %q", got.Path)
	}

	refresh, ok := byName["refresh-repo"]
	if !ok {
		t.Fatalf("skill refresh-repo not found; got %v", byName)
	}
	if refresh.Path != filepath.Join("openclaw-hub", "skills", "refresh-repo", "SKILL.md") {
		t.Errorf("path = %q", refresh.Path)
	}
}

func TestLoadSkill(t *testing.T) {
	root := t.TempDir()
	body := "---\nname: push-changes\ndescription: 'Publishes commits.'\n---\n\n# Push\n"
	write(t, filepath.Join(root, "pi-hub", "skills", "push-changes", "SKILL.md"), body)

	got, err := LoadSkill(root, "push-changes")
	if err != nil {
		t.Fatal(err)
	}
	if got != body {
		t.Errorf("LoadSkill = %q, want full body", got)
	}

	if _, err := LoadSkill(root, "../secrets"); err == nil {
		t.Error("path traversal must be rejected")
	}
	if _, err := LoadSkill(root, "missing"); err == nil {
		t.Error("missing skill must be an error")
	}
}

func TestScanMemoriesMissingDir(t *testing.T) {
	// The memories directory is optional: a checkout without it lists nothing.
	memories, err := ScanMemories(t.TempDir())
	if err != nil {
		t.Fatalf("missing memories dir must not error: %v", err)
	}
	if len(memories) != 0 {
		t.Errorf("want no memories, got %v", memories)
	}
}

func TestScanMemories(t *testing.T) {
	root := t.TempDir()
	write(t, filepath.Join(root, ".github", "memories", "conventions.md"),
		"---\nname: conventions\ndescription: 'House conventions.'\n---\n\n# Conventions\n")
	write(t, filepath.Join(root, "pi-hub", "memories", "incidents", "MEMORY.md"),
		"---\ndescription: Past incidents.\n---\n# Incidents\n")

	memories, err := ScanMemories(root)
	if err != nil {
		t.Fatal(err)
	}
	byName := map[string]Artifact{}
	for _, m := range memories {
		byName[m.Name] = m
	}

	got, ok := byName["conventions"]
	if !ok {
		t.Fatalf("memory conventions not found; got %v", byName)
	}
	if got.Description != "House conventions." {
		t.Errorf("description = %q", got.Description)
	}

	incident, ok := byName["incidents"]
	if !ok {
		t.Fatalf("memory incidents not found; got %v", byName)
	}
	if incident.Path != filepath.Join("pi-hub", "memories", "incidents", "MEMORY.md") {
		t.Errorf("path = %q", incident.Path)
	}
}

func TestLoadMemory(t *testing.T) {
	root := t.TempDir()
	body := "---\nname: conventions\ndescription: 'House conventions.'\n---\n\n# Conventions\n"
	write(t, filepath.Join(root, "memories", "conventions.md"), body)

	got, err := LoadMemory(root, "conventions")
	if err != nil {
		t.Fatal(err)
	}
	if got != body {
		t.Errorf("LoadMemory = %q, want full body", got)
	}

	if _, err := LoadMemory(root, "../secrets"); err == nil {
		t.Error("path traversal must be rejected")
	}
	if _, err := LoadMemory(root, "missing"); err == nil {
		t.Error("missing memory must be an error")
	}
}

func TestParseFrontmatter(t *testing.T) {
	cases := []struct {
		name string
		body string
		want Frontmatter
	}{
		{
			name: "single quotes",
			body: "---\nname: 'a'\ndescription: 'has: a colon'\n---\nbody",
			want: Frontmatter{Name: "a", Description: "has: a colon"},
		},
		{
			name: "double quotes",
			body: "---\nname: \"b\"\ndescription: \"plain\"\n---\n",
			want: Frontmatter{Name: "b", Description: "plain"},
		},
		{
			name: "no frontmatter",
			body: "# just markdown\n",
			want: Frontmatter{},
		},
		{
			name: "extra keys ignored",
			body: "---\nname: c\ntools: ['read']\ndescription: ok\n---\n",
			want: Frontmatter{Name: "c", Description: "ok"},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := ParseFrontmatter(tc.body); got != tc.want {
				t.Errorf("ParseFrontmatter = %+v, want %+v", got, tc.want)
			}
		})
	}
}
