package balancer

import (
	"fmt"
	"sync"
)

// Pool manages the active load balancer and supports hot-swapping algorithms
// while preserving backend state (request counts, alive status, etc.).
type Pool struct {
	mu        sync.RWMutex
	balancer  Balancer
	algorithm string
}

// NewPool creates a pool with the given algorithm and no backends.
func NewPool(algorithm string) (*Pool, error) {
	b, err := New(algorithm)
	if err != nil {
		return nil, err
	}
	return &Pool{balancer: b, algorithm: algorithm}, nil
}

// Balancer returns the currently active load balancer.
func (p *Pool) Balancer() Balancer {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.balancer
}

// Algorithm returns the name of the active algorithm.
func (p *Pool) Algorithm() string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.algorithm
}

// AddBackend registers a backend on the active balancer.
func (p *Pool) AddBackend(b *Backend) {
	p.Balancer().AddBackend(b)
}

// RemoveBackend removes a backend by URL from the active balancer.
func (p *Pool) RemoveBackend(url string) {
	p.Balancer().RemoveBackend(url)
}

// GetBackends returns all backends from the active balancer.
func (p *Pool) GetBackends() []*Backend {
	return p.Balancer().GetBackends()
}

// MarkAlive updates alive status on the active balancer.
func (p *Pool) MarkAlive(url string, alive bool) {
	p.Balancer().MarkAlive(url, alive)
}

// SetAlgorithm hot-swaps the load balancing algorithm, migrating existing backends.
func (p *Pool) SetAlgorithm(algorithm string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if algorithm == p.algorithm {
		return nil
	}

	newBal, err := New(algorithm)
	if err != nil {
		return fmt.Errorf("invalid algorithm: %w", err)
	}

	backends := p.balancer.GetBackends()
	for _, b := range backends {
		alive := b.Alive()
		newBal.AddBackend(b)
		newBal.MarkAlive(b.URL, alive)
	}

	p.balancer = newBal
	p.algorithm = algorithm
	return nil
}
