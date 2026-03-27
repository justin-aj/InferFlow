package metrics

import "sync/atomic"

type State struct {
	inFlight atomic.Int64
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
