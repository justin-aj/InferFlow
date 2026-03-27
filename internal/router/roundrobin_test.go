package router

import "testing"

func TestRoundRobinDistributesAcrossHealthyBackends(t *testing.T) {
	a, _ := NewBackend("a", "http://a.test")
	b, _ := NewBackend("b", "http://b.test")
	rr := NewRoundRobin([]*Backend{a, b})

	first, err := rr.Pick()
	if err != nil {
		t.Fatalf("pick first: %v", err)
	}
	second, err := rr.Pick()
	if err != nil {
		t.Fatalf("pick second: %v", err)
	}

	if first.Name == second.Name {
		t.Fatalf("expected different backends, got %s twice", first.Name)
	}
}

func TestRoundRobinSkipsUnhealthyBackends(t *testing.T) {
	a, _ := NewBackend("a", "http://a.test")
	b, _ := NewBackend("b", "http://b.test")
	a.SetHealthy(false)
	rr := NewRoundRobin([]*Backend{a, b})

	got, err := rr.Pick()
	if err != nil {
		t.Fatalf("pick: %v", err)
	}
	if got.Name != "b" {
		t.Fatalf("expected backend b, got %s", got.Name)
	}
}

func TestRoundRobinReturnsErrorWhenNoneHealthy(t *testing.T) {
	a, _ := NewBackend("a", "http://a.test")
	a.SetHealthy(false)
	rr := NewRoundRobin([]*Backend{a})

	if _, err := rr.Pick(); err != ErrNoHealthyBackend {
		t.Fatalf("expected ErrNoHealthyBackend, got %v", err)
	}
}
