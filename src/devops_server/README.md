# devops_server

MCP server written in Go. It reads the AI framework from a repository checkout
and exposes its **agents**, **prompts**, **skills** and **memories** as MCP
resources, with `list_agents`, `list_prompts`, `list_skills` and
`list_memories` tools for discovery. Built on the official
[`modelcontextprotocol/go-sdk`](https://github.com/modelcontextprotocol/go-sdk).

Agents are read from the external `project_agent_hub` checkout, reached through
the `./docs` symlink inside this module (see `FRAMEWORK_ROOT` below), so the
repository owns the content and this server just serves it.

There is **no per-artifact registration at startup**: artifacts are served
through URI templates, `framework://agents/{name}`, `framework://prompts/{name}`,
`framework://skills/{name}` and `framework://memories/{name}`, so content and the
set of available names never go stale — no restart needed after edits. Use the
`list_agents`, `list_prompts`, `list_skills` and `list_memories` tools to see
which names the templates accept.

## Requirements

- Go **1.25+** (the repo pins it via `.mise.toml`; run `mise install` if you use mise)
- Node.js/npm only for MCP Inspector

## Local setup

```bash
cd src/devops_server

# Optional: local overrides (transport, port, framework root, log level)
cp .env.example .env

make tidy   # sync dependencies
make build  # compile to ./bin/devops-server
```

## Run

```bash
make run        # stdio (default, for agents that spawn the server)
make run-http   # streamable HTTP on MCP_ADDR (default :8000)
```

Run the tests (in-memory transport smoke test of agent resources and tools):

```bash
make test
```

## Test with MCP Inspector

```bash
npx @modelcontextprotocol/inspector
```

Then in the Inspector UI:

- **stdio** — Transport `STDIO`, Command `go`, Args `run ./cmd/server`,
  working directory `src/devops_server`. Or point Command at `./bin/devops-server`.
- **HTTP** — start `make run-http` first, then Transport `Streamable HTTP`,
  URL `http://localhost:8000/mcp`.

Once connected, use **Tools → `list_agents`** / **`list_prompts`** /
**`list_skills`** / **`list_memories`** to discover names, then read any of them
as the resources `framework://agents/<name>`, `framework://prompts/<name>`,
`framework://skills/<name>` and `framework://memories/<name>` (all templates are
listed under **Resources → Templates**).

## Configuration

All variables are optional and documented in `.env.example`:
`MCP_TRANSPORT` (`stdio`|`http`), `MCP_ADDR`, `FRAMEWORK_ROOT`, `LOG_LEVEL`.
Logs go to stderr so they never corrupt the stdio transport.

`FRAMEWORK_ROOT` defaults to the `./docs` symlink inside this module, which
points at the framework's `.github` checkout (a leading `~` is expanded to your
home directory). It must point at a directory containing `agents/` (or
`.github/agents/`). Prompts are read from `prompts/` (or `.github/prompts/`),
skills from `<name>/SKILL.md` directories under `skills/` (or `.github/skills/`),
and memories from `memories/` (or `.github/memories/`), including any hub's
`*/skills/` and `*/memories/`. The `memories/` directory is optional.
