package router

import (
	"errors"
	"net/url"
	"strings"
	"sync"
)

var ErrNoHealthyBackend = errors.New("no healthy backend available")

type Backend struct {
	Name    string
	BaseURL string

	mu      sync.RWMutex
	healthy bool
}

func NewBackend(name, baseURL string) (*Backend, error) {
	if _, err := url.ParseRequestURI(baseURL); err != nil {
		return nil, err
	}

	return &Backend{
		Name:    strings.TrimSpace(name),
		BaseURL: strings.TrimRight(baseURL, "/"),
		healthy: true,
	}, nil
}

func (b *Backend) SetHealthy(healthy bool) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.healthy = healthy
}

func (b *Backend) Healthy() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.healthy
}
