package router

import (
	"context"
	"sync"
)

const (
	StrategyRoundRobin   = "round_robin"
	StrategyLeastPending = "least_pending"
	StrategyRandom       = "random"
	StrategyKVAware      = "kv_aware"
	StrategyCostAware    = "cost_aware"
)

type SelectionInput struct {
	Context       context.Context
	EstimatedCost int
	CacheKey      string
}

type Strategy interface {
	Name() string
	SetBackends(backends []*Backend)
	Select(input SelectionInput) (Decision, error)
	HasHealthyBackend() bool
}

type Decision struct {
	Backend         *Backend
	PendingRequests int64
	PendingCost     int64
	CacheHit        bool

	release func()
	once    sync.Once
}

func (d *Decision) Release() {
	d.once.Do(func() {
		if d.release != nil {
			d.release()
		}
	})
}
