package balancer

import (
	"sync/atomic"
	"testing"
)

func TestRoundRobin(t *testing.T) {
	rr := NewRoundRobin()
	b1 := &Backend{URL: "http://b1", Name: "b1", Weight: 1}
	b2 := &Backend{URL: "http://b2", Name: "b2", Weight: 1}
	rr.AddBackend(b1)
	rr.AddBackend(b2)

	seen := make(map[string]int)
	for i := 0; i < 100; i++ {
		b, err := rr.Next("1.2.3.4")
		if err != nil {
			t.Fatal(err)
		}
		seen[b.URL]++
	}

	if seen["http://b1"] != 50 || seen["http://b2"] != 50 {
		t.Errorf("expected even distribution, got %v", seen)
	}
}

func TestRoundRobinSkipsDead(t *testing.T) {
	rr := NewRoundRobin()
	b1 := &Backend{URL: "http://b1", Name: "b1", Weight: 1}
	b2 := &Backend{URL: "http://b2", Name: "b2", Weight: 1}
	rr.AddBackend(b1)
	rr.AddBackend(b2)
	b1.SetAlive(false)

	for i := 0; i < 10; i++ {
		b, err := rr.Next("1.2.3.4")
		if err != nil {
			t.Fatal(err)
		}
		if b.URL != "http://b2" {
			t.Errorf("expected only b2, got %s", b.URL)
		}
	}
}

func TestWeightedRoundRobin(t *testing.T) {
	wrr := NewWeightedRoundRobin()
	b1 := &Backend{URL: "http://b1", Name: "b1", Weight: 1}
	b2 := &Backend{URL: "http://b2", Name: "b2", Weight: 2}
	wrr.AddBackend(b1)
	wrr.AddBackend(b2)

	seen := make(map[string]int)
	for i := 0; i < 300; i++ {
		b, err := wrr.Next("")
		if err != nil {
			t.Fatal(err)
		}
		seen[b.URL]++
	}

	if seen["http://b1"] != 100 || seen["http://b2"] != 200 {
		t.Errorf("expected 1:2 ratio, got %v", seen)
	}
}

func TestLeastConnections(t *testing.T) {
	lc := NewLeastConnections()
	b1 := &Backend{URL: "http://b1", Name: "b1", Weight: 1}
	b2 := &Backend{URL: "http://b2", Name: "b2", Weight: 1}
	lc.AddBackend(b1)
	lc.AddBackend(b2)

	atomic.StoreInt64(&b1.ActiveConns, 5)
	atomic.StoreInt64(&b2.ActiveConns, 1)

	b, err := lc.Next("")
	if err != nil {
		t.Fatal(err)
	}
	if b.URL != "http://b2" {
		t.Errorf("expected b2 with fewer connections, got %s", b.URL)
	}
}

func TestIPHashSticky(t *testing.T) {
	ih := NewIPHash()
	b1 := &Backend{URL: "http://b1", Name: "b1", Weight: 1}
	b2 := &Backend{URL: "http://b2", Name: "b2", Weight: 1}
	ih.AddBackend(b1)
	ih.AddBackend(b2)

	first, err := ih.Next("10.0.0.1")
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 20; i++ {
		b, err := ih.Next("10.0.0.1")
		if err != nil {
			t.Fatal(err)
		}
		if b.URL != first.URL {
			t.Errorf("IP hash not sticky: got %s want %s", b.URL, first.URL)
		}
	}
}

func TestNoAliveBackends(t *testing.T) {
	tests := []struct {
		name string
		b    Balancer
	}{
		{"round_robin", NewRoundRobin()},
		{"weighted", NewWeightedRoundRobin()},
		{"least_conn", NewLeastConnections()},
		{"ip_hash", NewIPHash()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &Backend{URL: "http://b1", Name: "b1", Weight: 1}
			tt.b.AddBackend(b)
			b.SetAlive(false)
			_, err := tt.b.Next("1.2.3.4")
			if err == nil {
				t.Error("expected error when no alive backends")
			}
		})
	}
}
