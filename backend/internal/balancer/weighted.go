package balancer

import (
	"fmt"
	"sync"
	"sync/atomic"
)

// WeightedRoundRobin distributes requests proportionally to backend weights.
type WeightedRoundRobin struct {
	mu       sync.RWMutex
	backends []*Backend
	slots    []*Backend
	counter  uint64
}

// NewWeightedRoundRobin creates a new weighted round-robin load balancer.
func NewWeightedRoundRobin() *WeightedRoundRobin {
	return &WeightedRoundRobin{}
}

func (w *WeightedRoundRobin) rebuildSlots() {
	var slots []*Backend
	for _, b := range w.backends {
		if !b.Alive() {
			continue
		}
		for i := 0; i < b.Weight; i++ {
			slots = append(slots, b)
		}
	}
	w.slots = slots
}

// Next returns the next backend using weighted round-robin.
func (w *WeightedRoundRobin) Next(_ string) (*Backend, error) {
	w.mu.RLock()
	defer w.mu.RUnlock()

	var alive []*Backend
	for _, b := range w.slots {
		if b.Alive() {
			alive = append(alive, b)
		}
	}
	if len(alive) == 0 {
		return nil, fmt.Errorf("no alive backends available")
	}

	idx := atomic.AddUint64(&w.counter, 1) - 1
	return alive[idx%uint64(len(alive))], nil
}

// AddBackend adds a backend and rebuilds weight slots.
func (w *WeightedRoundRobin) AddBackend(b *Backend) {
	w.mu.Lock()
	defer w.mu.Unlock()
	b.SetAlive(true)
	w.backends = append(w.backends, b)
	w.rebuildSlots()
}

// RemoveBackend removes a backend by URL.
func (w *WeightedRoundRobin) RemoveBackend(url string) {
	w.mu.Lock()
	defer w.mu.Unlock()
	for i, b := range w.backends {
		if b.URL == url {
			w.backends = append(w.backends[:i], w.backends[i+1:]...)
			w.rebuildSlots()
			return
		}
	}
}

// GetBackends returns all registered backends.
func (w *WeightedRoundRobin) GetBackends() []*Backend {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return cloneBackends(w.backends)
}

// MarkAlive updates alive status and rebuilds slots.
func (w *WeightedRoundRobin) MarkAlive(url string, alive bool) {
	w.mu.Lock()
	defer w.mu.Unlock()
	for _, b := range w.backends {
		if b.URL == url {
			b.SetAlive(alive)
			w.rebuildSlots()
			return
		}
	}
}
