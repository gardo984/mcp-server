// Package framework reads the AI framework from a repository checkout. It
// exposes four artifact kinds, agents, prompts, skills and memories, stored as
//
//	<root>/.github/agents/<name>.agent.md
//	<root>/.github/prompts/<name>.prompt.md
//	<root>/.github/skills/<name>.prompt.md
//	<root>/.github/memories/<name>.md
//
// with the flat <root>/agents/<name>.md, <root>/prompts/<name>.md,
// <root>/skills/<name>.md and <root>/memories/<name>.md layouts supported as
// fallbacks. Skills and memories may also be directories holding a canonical
// entry file (<name>/SKILL.md, <name>/MEMORY.md); those directories are found
// under any hub, i.e. <root>/skills and <root>/*/skills, so a multi-hub checkout
// works without configuration. Missing directories are always skipped.
//
// The package is deliberately transport-agnostic: it has no dependency on the
// MCP SDK, so it can be unit-tested and reused by other front ends.
package framework

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Artifact is a single framework file discovered in the repository.
type Artifact struct {
	Name        string // derived from the file name, e.g. "backend-sre"
	Path        string // path relative to the framework root
	Description string // frontmatter description, when present
}

// kindSpec describes where a class of artifacts lives under the framework root
// and how their names are derived from file names.
type kindSpec struct {
	// dirs are candidate directories, relative to the framework root, in
	// priority order. `.github/<kind>` wins over a flat `<kind>` directory, so a
	// GitHub-style checkout works out of the box.
	dirs []string
	// suffixes identify artifact files. The longest matching suffix wins, so
	// "backend-sre.agent.md" yields "backend-sre" and not "backend-sre.agent".
	suffixes []string
	// entryFile, when set, marks directory-based artifacts: a sub-directory
	// <name> holding entryFile (e.g. skills/<name>/SKILL.md) is an artifact
	// named after the directory. Its file name never contributes to the name.
	entryFile string
}

var agentsKind = kindSpec{
	dirs:     []string{filepath.Join(".github", "agents"), "agents"},
	suffixes: []string{".agent.md", ".md", ".markdown", ".txt"},
}

var promptsKind = kindSpec{
	dirs:     []string{filepath.Join(".github", "prompts"), "prompts"},
	suffixes: []string{".prompt.md", ".md", ".markdown", ".txt"},
}

// skillsKind covers both flat skill files and the directory-per-skill layout.
// The "*/skills" entry picks up every hub (pi-hub, openclaw-hub, ...) without
// hardcoding their names; ".github/skills" wins and de-duplicates by name.
var skillsKind = kindSpec{
	dirs:      []string{filepath.Join(".github", "skills"), "skills", filepath.Join("*", "skills")},
	suffixes:  []string{".skill.md", ".prompt.md", ".md", ".markdown", ".txt"},
	entryFile: "SKILL.md",
}

// memoriesKind mirrors skillsKind for durable notes. The directory is optional:
// a checkout without it simply lists no memories.
var memoriesKind = kindSpec{
	dirs:      []string{filepath.Join(".github", "memories"), "memories", filepath.Join("*", "memories")},
	suffixes:  []string{".memory.md", ".md", ".markdown", ".txt"},
	entryFile: "MEMORY.md",
}

// ScanAgents walks the agent directories under root and returns the agents it
// finds. Missing directories are skipped rather than reported as errors.
func ScanAgents(root string) ([]Artifact, error) { return scan(root, agentsKind) }

// ScanPrompts walks the prompt directories under root and returns the prompts
// it finds. Missing directories are skipped rather than reported as errors.
func ScanPrompts(root string) ([]Artifact, error) { return scan(root, promptsKind) }

// ScanSkills walks the skill directories under root and returns the skills it
// finds. Missing directories are skipped rather than reported as errors.
func ScanSkills(root string) ([]Artifact, error) { return scan(root, skillsKind) }

// ScanMemories walks the memory directories under root and returns the memories
// it finds. Missing directories are skipped rather than reported as errors.
func ScanMemories(root string) ([]Artifact, error) { return scan(root, memoriesKind) }

// LoadAgent reads one agent by name. It rejects any name that is not a plain
// file name, so it can never escape the framework root (path traversal).
func LoadAgent(root, name string) (string, error) { return load(root, name, agentsKind) }

// LoadPrompt reads one prompt by name. It rejects any name that is not a plain
// file name, so it can never escape the framework root (path traversal).
func LoadPrompt(root, name string) (string, error) { return load(root, name, promptsKind) }

// LoadSkill reads one skill by name. It rejects any name that is not a plain
// file name, so it can never escape the framework root (path traversal).
func LoadSkill(root, name string) (string, error) { return load(root, name, skillsKind) }

// LoadMemory reads one memory by name. It rejects any name that is not a plain
// file name, so it can never escape the framework root (path traversal).
func LoadMemory(root, name string) (string, error) { return load(root, name, memoriesKind) }

// scan collects the artifacts of one kind under root, de-duplicating by name so
// a lower-priority directory cannot shadow a higher-priority one.
func scan(root string, spec kindSpec) ([]Artifact, error) {
	dirs, err := expandDirs(root, spec.dirs)
	if err != nil {
		return nil, err
	}

	var items []Artifact
	seen := make(map[string]bool)

	for _, dir := range dirs {
		entries, err := os.ReadDir(filepath.Join(root, dir))
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, fmt.Errorf("scan %s: %w", dir, err)
		}
		for _, entry := range entries {
			// Skip dotfiles (e.g. .gitkeep) and hidden directories.
			if strings.HasPrefix(entry.Name(), ".") {
				continue
			}
			name, path, ok := artifactPath(root, dir, entry, spec)
			if !ok {
				continue
			}
			if seen[name] {
				continue
			}
			seen[name] = true

			item := Artifact{Name: name, Path: path}
			// Best effort: read errors surface on Load, not during Scan.
			if body, err := os.ReadFile(filepath.Join(root, path)); err == nil {
				item.Description = ParseFrontmatter(string(body)).Description
			}
			items = append(items, item)
		}
	}

	sort.Slice(items, func(i, j int) bool { return items[i].Name < items[j].Name })
	return items, nil
}

// artifactPath reports whether a directory entry is an artifact of the kind and,
// if so, returns its name and its path relative to the framework root. Two
// layouts are supported: directories holding spec.entryFile (<name>/SKILL.md)
// and flat files whose name ends in one of spec.suffixes.
func artifactPath(root, dir string, entry os.DirEntry, spec kindSpec) (string, string, bool) {
	full := filepath.Join(dir, entry.Name())

	if spec.entryFile != "" && isDir(filepath.Join(root, full)) {
		entryPath := filepath.Join(full, spec.entryFile)
		if _, err := os.Stat(filepath.Join(root, entryPath)); err == nil {
			return entry.Name(), entryPath, true
		}
		return "", "", false
	}

	if entry.IsDir() {
		return "", "", false
	}
	name, ok := artifactName(entry.Name(), spec.suffixes)
	if !ok {
		return "", "", false
	}
	return name, full, true
}

// load reads one artifact of the given kind by name.
func load(root, name string, spec kindSpec) (string, error) {
	if !validName(name) {
		return "", fmt.Errorf("invalid name %q", name)
	}
	dirs, err := expandDirs(root, spec.dirs)
	if err != nil {
		return "", err
	}
	for _, dir := range dirs {
		for _, candidate := range candidates(dir, name, spec) {
			body, err := os.ReadFile(filepath.Join(root, candidate))
			if err == nil {
				return string(body), nil
			}
			if !os.IsNotExist(err) {
				return "", fmt.Errorf("read %s: %w", candidate, err)
			}
		}
	}
	return "", fmt.Errorf("framework item %q not found under %s", name, root)
}

// expandDirs resolves the candidate directories of a kind against root,
// expanding glob patterns (e.g. "*/skills") so nested hubs are discovered
// without configuration. Order is preserved and duplicates are dropped, which
// keeps the priority encoded in kindSpec.dirs.
func expandDirs(root string, dirs []string) ([]string, error) {
	var out []string
	seen := make(map[string]bool)
	for _, dir := range dirs {
		matches, err := filepath.Glob(filepath.Join(root, dir))
		if err != nil {
			return nil, fmt.Errorf("expand %s: %w", dir, err)
		}
		for _, match := range matches {
			rel, err := filepath.Rel(root, match)
			if err != nil || rel == "." || strings.HasPrefix(rel, "..") {
				continue
			}
			if seen[rel] {
				continue
			}
			seen[rel] = true
			out = append(out, rel)
		}
	}
	return out, nil
}

// isDir reports whether path is a directory, following symlinks.
func isDir(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

// ParseFrontmatter extracts the leading YAML frontmatter block, if any. It is a
// deliberately small parser: only the flat, single-line keys the framework uses
// are understood, which avoids pulling in a YAML dependency.
func ParseFrontmatter(body string) Frontmatter {
	var fm Frontmatter
	block, ok := frontmatterBlock(body)
	if !ok {
		return fm
	}
	for _, line := range strings.Split(block, "\n") {
		key, value, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		value = unquote(strings.TrimSpace(value))
		switch strings.TrimSpace(key) {
		case "name":
			fm.Name = value
		case "description":
			fm.Description = value
		}
	}
	return fm
}

// Frontmatter is the subset of an artifact's YAML frontmatter the framework
// uses. Unknown keys are ignored.
type Frontmatter struct {
	Name        string
	Description string
}

// candidates lists the file paths to try for one artifact, most specific first.
func candidates(dir, name string, spec kindSpec) []string {
	out := make([]string, 0, len(spec.suffixes)+2)
	if spec.entryFile != "" {
		out = append(out, filepath.Join(dir, name, spec.entryFile))
	}
	for _, suffix := range spec.suffixes {
		out = append(out, filepath.Join(dir, name+suffix))
	}
	out = append(out, filepath.Join(dir, name))
	return out
}

// artifactName derives the artifact name from a file name.
func artifactName(fileName string, suffixes []string) (string, bool) {
	best := ""
	for _, suffix := range suffixes {
		if strings.HasSuffix(fileName, suffix) && len(suffix) > len(best) {
			best = suffix
		}
	}
	if best == "" {
		return "", false
	}
	name := strings.TrimSuffix(fileName, best)
	if name == "" {
		return "", false
	}
	return name, true
}

// frontmatterBlock returns the text between the opening and closing "---"
// fences, if the body starts with one.
func frontmatterBlock(body string) (string, bool) {
	if !strings.HasPrefix(body, "---") {
		return "", false
	}
	rest := strings.TrimPrefix(body, "---")
	rest = strings.TrimPrefix(rest, "\r")
	rest = strings.TrimPrefix(rest, "\n")
	end := strings.Index(rest, "\n---")
	if end < 0 {
		return "", false
	}
	return rest[:end], true
}

// unquote strips a single pair of surrounding single or double quotes.
func unquote(s string) string {
	if len(s) >= 2 {
		if (s[0] == '\'' && s[len(s)-1] == '\'') || (s[0] == '"' && s[len(s)-1] == '"') {
			return s[1 : len(s)-1]
		}
	}
	return s
}

// validName allows a plain file name only: no separators, no parent refs.
func validName(name string) bool {
	if name == "" || name == "." || name == ".." {
		return false
	}
	if name != filepath.Base(name) {
		return false
	}
	return !strings.ContainsAny(name, `/\`)
}
