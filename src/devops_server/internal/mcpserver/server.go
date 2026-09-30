// Package mcpserver assembles the MCP server: it registers the framework
// resources and tools and owns the transport lifecycle.
package mcpserver

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/gardo984/mcp-server/src/devops_server/internal/config"
	"github.com/gardo984/mcp-server/src/devops_server/internal/resources"
	"github.com/gardo984/mcp-server/src/devops_server/internal/tools"
)

const (
	serverName = "devops-mcp"
	version    = "0.1.0"
)

// Server owns the MCP server and its configuration.
type Server struct {
	mcp *mcp.Server
	cfg config.Config
	log *slog.Logger
}

// New builds the MCP server and registers all capabilities.
func New(cfg config.Config, log *slog.Logger) (*Server, error) {
	s := mcp.NewServer(&mcp.Implementation{
		Name:        serverName,
		Title:       "DevOps MCP",
		Version:     version,
		Description: "Exposes the AI framework agents, prompts, skills and memories from a repository as MCP resources, with list_agents, list_prompts, list_skills and list_memories tools for discovery.",
	}, nil)

	tools.Register(s, cfg.FrameworkRoot)
	resources.Register(s, cfg.FrameworkRoot)

	log.Info("agent capabilities registered", "framework_root", cfg.FrameworkRoot)
	return &Server{mcp: s, cfg: cfg, log: log}, nil
}

// RunStdio serves MCP over stdin/stdout. This is the default when an MCP host
// spawns the server as a child process.
func (s *Server) RunStdio(ctx context.Context) error {
	s.log.Info("serving MCP over stdio")
	return s.mcp.Run(ctx, &mcp.StdioTransport{})
}

// RunHTTP serves MCP over the streamable HTTP transport until ctx is done.
func (s *Server) RunHTTP(ctx context.Context) error {
	handler := mcp.NewStreamableHTTPHandler(
		func(*http.Request) *mcp.Server { return s.mcp },
		nil,
	)
	srv := &http.Server{
		Addr:              s.cfg.Addr,
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			s.log.Error("http shutdown", "err", err)
		}
	}()

	s.log.Info("serving MCP over http", "addr", s.cfg.Addr)
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}
