package balancer

import (
	"fmt"
	"sync/atomic"
	"time"
)

// Backend represents an upstream server in the load balancer pool.
type Backend struct {
	URL           string
	Name          string
	Weight        int
	ActiveConns   int64
	IsAlive       int64 // 1 = alive, 0 = dead
	TotalRequests int64
	TotalErrors   int64
	LastCheck     atomic.Value // time.Time
}

// Alive returns whether the backend is currently marked alive.
func (b *Backend) Alive() bool {
	return atomic.LoadInt64(&b.IsAlive) == 1
}

// SetAlive sets the alive status of the backend.
func (b *Backend) SetAlive(alive bool) {
	if alive {
		atomic.StoreInt64(&b.IsAlive, 1)
	} else {
		atomic.StoreInt64(&b.IsAlive, 0)
	}
	b.LastCheck.Store(time.Now())
}

// Balancer selects the next backend for an incoming request.
type Balancer interface {
	Next(clientIP string) (*Backend, error)
	AddBackend(b *Backend)
	RemoveBackend(url string)
	GetBackends() []*Backend
	MarkAlive(url string, alive bool)
}

// New creates a Balancer for the given algorithm name.
func New(algorithm string) (Balancer, error) {
	switch algorithm {
	case "round_robin":
		return NewRoundRobin(), nil
	case "weighted_round_robin":
		return NewWeightedRoundRobin(), nil
	case "least_connections":
		return NewLeastConnections(), nil
	case "ip_hash":
		return NewIPHash(), nil
	default:
		return nil, fmt.Errorf("unknown algorithm: %s", algorithm)
	}
}

// cloneBackends returns a shallow copy of the backend slice.
func cloneBackends(backends []*Backend) []*Backend {
	out := make([]*Backend, len(backends))
	copy(out, backends)
	return out
}

// aliveBackends returns only backends marked as alive.
func aliveBackends(backends []*Backend) []*Backend {
	var alive []*Backend
	for _, b := range backends {
		if b.Alive() {
			alive = append(alive, b)
		}
	}
	return alive
}
