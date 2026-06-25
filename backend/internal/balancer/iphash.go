package balancer

import (
	"fmt"
	"hash/fnv"
	"sync"
)

// IPHash routes the same client IP to the same backend (sticky sessions).
type IPHash struct {
	mu       sync.RWMutex
	backends []*Backend
}

// NewIPHash creates a new IP-hash load balancer.
func NewIPHash() *IPHash {
	return &IPHash{}
}

// Next returns a backend selected by hashing the client IP.
func (i *IPHash) Next(clientIP string) (*Backend, error) {
	i.mu.RLock()
	alive := aliveBackends(i.backends)
	i.mu.RUnlock()

	if len(alive) == 0 {
		return nil, fmt.Errorf("no alive backends available")
	}

	h := fnv.New32a()
	_, _ = h.Write([]byte(clientIP))
	idx := h.Sum32() % uint32(len(alive))
	return alive[idx], nil
}

// AddBackend adds a backend to the pool.
func (i *IPHash) AddBackend(b *Backend) {
	i.mu.Lock()
	defer i.mu.Unlock()
	b.SetAlive(true)
	i.backends = append(i.backends, b)
}

// RemoveBackend removes a backend by URL.
func (i *IPHash) RemoveBackend(url string) {
	i.mu.Lock()
	defer i.mu.Unlock()
	for idx, b := range i.backends {
		if b.URL == url {
			i.backends = append(i.backends[:idx], i.backends[idx+1:]...)
			return
		}
	}
}

// GetBackends returns all registered backends.
func (i *IPHash) GetBackends() []*Backend {
	i.mu.RLock()
	defer i.mu.RUnlock()
	return cloneBackends(i.backends)
}

// MarkAlive updates the alive status of a backend.
func (i *IPHash) MarkAlive(url string, alive bool) {
	i.mu.RLock()
	defer i.mu.RUnlock()
	for _, b := range i.backends {
		if b.URL == url {
			b.SetAlive(alive)
			return
		}
	}
}
