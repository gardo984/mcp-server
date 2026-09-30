// Command server runs the DevOps MCP server.
//
// Usage:
//
//	MCP_TRANSPORT=stdio go run ./cmd/server   # for locally spawned agents
//	MCP_TRANSPORT=http  go run ./cmd/server   # streamable HTTP on MCP_ADDR
//
// See .env.example for every supported environment variable.
package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/gardo984/mcp-server/src/devops_server/internal/config"
	"github.com/gardo984/mcp-server/src/devops_server/internal/mcpserver"
)

func main() {
	if err := run(); err != nil {
		slog.Error("fatal", "err", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	// Logs always go to stderr: stdout belongs to the stdio transport.
	log := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: cfg.SlogLevel(),
	}))

	srv, err := mcpserver.New(cfg, log)
	if err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if cfg.Transport == config.TransportHTTP {
		return srv.RunHTTP(ctx)
	}
	return srv.RunStdio(ctx)
}
