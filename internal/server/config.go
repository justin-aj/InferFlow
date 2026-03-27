package server

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/ajinfrank/inferflow/internal/router"
)

type Config struct {
	ListenAddr           string
	Backends             []*router.Backend
	ProbeInterval        time.Duration
	BackendRequestTimout time.Duration
}

func LoadConfigFromEnv() (Config, error) {
	listenAddr := getenv("INFERFLOW_LISTEN_ADDR", ":8080")
	backendURLs := splitAndTrim(getenv("INFERFLOW_BACKENDS", "http://localhost:9000"))
	if len(backendURLs) == 0 {
		return Config{}, fmt.Errorf("INFERFLOW_BACKENDS must contain at least one backend URL")
	}

	backends := make([]*router.Backend, 0, len(backendURLs))
	for idx, raw := range backendURLs {
		backend, err := router.NewBackend(fmt.Sprintf("backend-%d", idx+1), raw)
		if err != nil {
			return Config{}, fmt.Errorf("invalid backend %q: %w", raw, err)
		}
		backends = append(backends, backend)
	}

	return Config{
		ListenAddr:           listenAddr,
		Backends:             backends,
		ProbeInterval:        2 * time.Second,
		BackendRequestTimout: 10 * time.Second,
	}, nil
}

func getenv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func splitAndTrim(value string) []string {
	parts := strings.Split(value, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}
