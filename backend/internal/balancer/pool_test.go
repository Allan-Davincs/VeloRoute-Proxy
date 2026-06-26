package balancer

import "testing"

func TestPoolHotSwapPreservesBackends(t *testing.T) {
	pool, err := NewPool("round_robin")
	if err != nil {
		t.Fatal(err)
	}

	b1 := &Backend{URL: "http://b1", Name: "b1", Weight: 1}
	b2 := &Backend{URL: "http://b2", Name: "b2", Weight: 2}
	pool.AddBackend(b1)
	pool.AddBackend(b2)
	b1.SetAlive(false)

	if err := pool.SetAlgorithm("weighted_round_robin"); err != nil {
		t.Fatal(err)
	}

	if pool.Algorithm() != "weighted_round_robin" {
		t.Fatalf("expected weighted_round_robin, got %s", pool.Algorithm())
	}

	backends := pool.GetBackends()
	if len(backends) != 2 {
		t.Fatalf("expected 2 backends, got %d", len(backends))
	}
	if backends[0].Alive() {
		t.Fatal("expected b1 to remain dead after swap")
	}

	_, err = pool.Balancer().Next("")
	if err != nil {
		t.Fatal("expected request to succeed with one alive backend")
	}
}

func TestPoolInvalidAlgorithm(t *testing.T) {
	pool, _ := NewPool("round_robin")
	if err := pool.SetAlgorithm("invalid_algo"); err == nil {
		t.Fatal("expected error for invalid algorithm")
	}
}
