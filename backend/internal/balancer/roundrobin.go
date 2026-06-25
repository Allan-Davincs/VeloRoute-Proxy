package balancer

import (
	"fmt"
	"sync"
	"sync/atomic"
)

// RoundRobin distributes requests evenly across alive backends.
type RoundRobin struct {
	mu       sync.RWMutex
	backends []*Backend
	counter  uint64
}

// NewRoundRobin creates a new round-robin load balancer.
func NewRoundRobin() *RoundRobin {
	return &RoundRobin{}
}

// Next returns the next alive backend in round-robin order.
func (r *RoundRobin) Next(_ string) (*Backend, error) {
	r.mu.RLock()
	alive := aliveBackends(r.backends)
	r.mu.RUnlock()

	if len(alive) == 0 {
		return nil, fmt.Errorf("no alive backends available")
	}

	idx := atomic.AddUint64(&r.counter, 1) - 1
	return alive[idx%uint64(len(alive))], nil
}

// AddBackend adds a backend to the pool.
func (r *RoundRobin) AddBackend(b *Backend) {
	r.mu.Lock()
	defer r.mu.Unlock()
	b.SetAlive(true)
	r.backends = append(r.backends, b)
}

// RemoveBackend removes a backend by URL.
func (r *RoundRobin) RemoveBackend(url string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i, b := range r.backends {
		if b.URL == url {
			r.backends = append(r.backends[:i], r.backends[i+1:]...)
			return
		}
	}
}

// GetBackends returns all registered backends.
func (r *RoundRobin) GetBackends() []*Backend {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return cloneBackends(r.backends)
}

// MarkAlive updates the alive status of a backend.
func (r *RoundRobin) MarkAlive(url string, alive bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, b := range r.backends {
		if b.URL == url {
			b.SetAlive(alive)
			return
		}
	}
}
