package metrics

import (
	"sort"
	"sync"
	"sync/atomic"
)

type State struct {
	inFlight      atomic.Int64
	requestsTotal atomic.Int64
	backendErrors atomic.Int64
	kvCacheHits   atomic.Int64
	kvCacheMisses atomic.Int64

	mu             sync.RWMutex
	strategyCounts map[string]*atomic.Int64
	backendCounts  map[string]*atomic.Int64
	latencyEMA     map[string]float64 // milliseconds, exponential moving average
}

func (s *State) IncInFlight() {
	s.inFlight.Add(1)
}

func (s *State) DecInFlight() {
	s.inFlight.Add(-1)
}

func (s *State) InFlight() int64 {
	return s.inFlight.Load()
}

func (s *State) IncRequestsTotal() {
	s.requestsTotal.Add(1)
}

func (s *State) RequestsTotal() int64 {
	return s.requestsTotal.Load()
}

func (s *State) IncBackendErrors() {
	s.backendErrors.Add(1)
}

func (s *State) BackendErrors() int64 {
	return s.backendErrors.Load()
}

func (s *State) IncKVCacheHit()    { s.kvCacheHits.Add(1) }
func (s *State) IncKVCacheMiss()   { s.kvCacheMisses.Add(1) }
func (s *State) KVCacheHits() int64  { return s.kvCacheHits.Load() }
func (s *State) KVCacheMisses() int64 { return s.kvCacheMisses.Load() }

func (s *State) RecordLatency(backendName string, ms float64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.latencyEMA == nil {
		s.latencyEMA = make(map[string]float64)
	}
	const alpha = 0.2
	if prev, ok := s.latencyEMA[backendName]; ok {
		s.latencyEMA[backendName] = alpha*ms + (1-alpha)*prev
	} else {
		s.latencyEMA[backendName] = ms
	}
}

func (s *State) LatencySnapshot() map[string]int64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make(map[string]int64, len(s.latencyEMA))
	for k, v := range s.latencyEMA {
		out[k] = int64(v)
	}
	return out
}

func (s *State) RecordStrategy(name string) {
	s.counterFor(&s.strategyCounts, name).Add(1)
}

func (s *State) RecordBackend(name string) {
	s.counterFor(&s.backendCounts, name).Add(1)
}

func (s *State) StrategySnapshot() map[string]int64 {
	return s.snapshot(s.strategyCounts)
}

func (s *State) BackendSnapshot() map[string]int64 {
	return s.snapshot(s.backendCounts)
}

func (s *State) SortedKeys(values map[string]int64) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func (s *State) counterFor(target *map[string]*atomic.Int64, name string) *atomic.Int64 {
	s.mu.Lock()
	defer s.mu.Unlock()

	if *target == nil {
		*target = make(map[string]*atomic.Int64)
	}
	if counter, ok := (*target)[name]; ok {
		return counter
	}
	counter := &atomic.Int64{}
	(*target)[name] = counter
	return counter
}

func (s *State) snapshot(source map[string]*atomic.Int64) map[string]int64 {
	s.mu.RLock()
	defer s.mu.RUnlock()

	out := make(map[string]int64, len(source))
	for key, value := range source {
		out[key] = value.Load()
	}
	return out
}
