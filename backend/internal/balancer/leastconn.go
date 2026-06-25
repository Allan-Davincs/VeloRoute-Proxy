package balancer

import (
	"fmt"
	"sync"
	"sync/atomic"
)

// LeastConnections routes to the backend with the fewest active connections.
type LeastConnections struct {
	mu       sync.RWMutex
	backends []*Backend
}

// NewLeastConnections creates a new least-connections load balancer.
func NewLeastConnections() *LeastConnections {
	return &LeastConnections{}
}

// Next returns the alive backend with the lowest active connection count.
func (l *LeastConnections) Next(_ string) (*Backend, error) {
	l.mu.RLock()
	alive := aliveBackends(l.backends)
	l.mu.RUnlock()

	if len(alive) == 0 {
		return nil, fmt.Errorf("no alive backends available")
	}

	var selected *Backend
	var minConns int64 = -1
	for _, b := range alive {
		conns := atomic.LoadInt64(&b.ActiveConns)
		if minConns == -1 || conns < minConns {
			minConns = conns
			selected = b
		}
	}
	return selected, nil
}

// AddBackend adds a backend to the pool.
func (l *LeastConnections) AddBackend(b *Backend) {
	l.mu.Lock()
	defer l.mu.Unlock()
	b.SetAlive(true)
	l.backends = append(l.backends, b)
}

// RemoveBackend removes a backend by URL.
func (l *LeastConnections) RemoveBackend(url string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	for i, b := range l.backends {
		if b.URL == url {
			l.backends = append(l.backends[:i], l.backends[i+1:]...)
			return
		}
	}
}

// GetBackends returns all registered backends.
func (l *LeastConnections) GetBackends() []*Backend {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return cloneBackends(l.backends)
}

// MarkAlive updates the alive status of a backend.
func (l *LeastConnections) MarkAlive(url string, alive bool) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	for _, b := range l.backends {
		if b.URL == url {
			b.SetAlive(alive)
			return
		}
	}
}
