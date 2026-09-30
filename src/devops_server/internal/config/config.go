// Package config loads and validates the process configuration.
//
// Every knob is an environment variable so the same binary can run unchanged
// in a local shell, inside a container, or spawned by an MCP host. Defaults are
// safe for local development: stdio transport and the local AI framework
// checkout as the framework root.
package config

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
)

// Transport selects how the MCP protocol is exposed.
type Transport string

const (
	// TransportStdio communicates over stdin/stdout. This is what MCP hosts
	// expect when they spawn the server as a child process.
	TransportStdio Transport = "stdio"
	// TransportHTTP serves the streamable HTTP transport.
	TransportHTTP Transport = "http"
)

// Config is the fully resolved, validated configuration.
type Config struct {
	Transport     Transport
	Addr          string
	FrameworkRoot string
	LogLevel      string
}

// defaultFrameworkRoot is the AI framework repository this server exposes. It
// is overridden by FRAMEWORK_ROOT. It defaults to the "docs" symlink next to
// this module, which points at the framework's .github checkout. A leading "~"
// is expanded to the home directory.
const defaultFrameworkRoot = "./docs"

// Load reads the environment, applies defaults and validates the result.
func Load() (Config, error) {
	cfg := Config{
		Transport:     Transport(getenv("MCP_TRANSPORT", string(TransportStdio))),
		Addr:          getenv("MCP_ADDR", ":8000"),
		FrameworkRoot: expandHome(getenv("FRAMEWORK_ROOT", defaultFrameworkRoot)),
		LogLevel:      strings.ToLower(getenv("LOG_LEVEL", "info")),
	}
	return cfg, cfg.validate()
}

func (c Config) validate() error {
	switch c.Transport {
	case TransportStdio, TransportHTTP:
	default:
		return fmt.Errorf("invalid MCP_TRANSPORT %q: want %q or %q", c.Transport, TransportStdio, TransportHTTP)
	}
	switch c.LogLevel {
	case "debug", "info", "warn", "error":
	default:
		return fmt.Errorf("invalid LOG_LEVEL %q: want debug, info, warn or error", c.LogLevel)
	}
	if c.FrameworkRoot == "" {
		return fmt.Errorf("FRAMEWORK_ROOT must not be empty")
	}
	info, err := os.Stat(c.FrameworkRoot)
	if err != nil {
		return fmt.Errorf("FRAMEWORK_ROOT %q: %w", c.FrameworkRoot, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("FRAMEWORK_ROOT %q is not a directory", c.FrameworkRoot)
	}
	return nil
}

// SlogLevel maps the configured level onto slog.
func (c Config) SlogLevel() slog.Level {
	switch c.LogLevel {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

// expandHome expands a leading "~" or "~/" to the current user's home
// directory. Paths without a leading tilde are returned unchanged.
func expandHome(path string) string {
	if path != "~" && !strings.HasPrefix(path, "~/") {
		return path
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return path
	}
	if path == "~" {
		return home
	}
	return filepath.Join(home, path[2:])
}
