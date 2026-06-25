package health

import (
	"context"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/allan-davincs/veloroute/internal/balancer"
)

// Checker performs periodic health checks on backend servers.
type Checker struct {
	enabled  bool
	interval time.Duration
	timeout  time.Duration
	path     string
	client   *http.Client
	balancer balancer.Balancer
	logger   *slog.Logger
	stopCh   chan struct{}
	wg       sync.WaitGroup
}

// NewChecker creates a new health checker.
func NewChecker(enabled bool, intervalSec, timeoutSec int, path string, b balancer.Balancer, logger *slog.Logger) *Checker {
	return &Checker{
		enabled:  enabled,
		interval: time.Duration(intervalSec) * time.Second,
		timeout:  time.Duration(timeoutSec) * time.Second,
		path:     path,
		client: &http.Client{
			Timeout: time.Duration(timeoutSec) * time.Second,
		},
		balancer: b,
		logger:   logger,
		stopCh:   make(chan struct{}),
	}
}

// Start begins health checking for all backends.
func (c *Checker) Start() {
	if !c.enabled {
		for _, b := range c.balancer.GetBackends() {
			c.balancer.MarkAlive(b.URL, true)
		}
		return
	}

	for _, b := range c.balancer.GetBackends() {
		c.wg.Add(1)
		go c.checkLoop(b)
	}
}

func (c *Checker) checkLoop(b *balancer.Backend) {
	defer c.wg.Done()
	ticker := time.NewTicker(c.interval)
	defer ticker.Stop()

	c.check(b)

	for {
		select {
		case <-ticker.C:
			c.check(b)
		case <-c.stopCh:
			return
		}
	}
}

func (c *Checker) check(b *balancer.Backend) {
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, b.URL+c.path, nil)
	if err != nil {
		c.transition(b, false)
		return
	}

	resp, err := c.client.Do(req)
	if err != nil {
		c.transition(b, false)
		return
	}
	defer resp.Body.Close()

	alive := resp.StatusCode >= 200 && resp.StatusCode < 400
	c.transition(b, alive)
}

func (c *Checker) transition(b *balancer.Backend, alive bool) {
	wasAlive := b.Alive()
	c.balancer.MarkAlive(b.URL, alive)

	if wasAlive && !alive {
		c.logger.Warn("Backend is DOWN", "backend", b.Name, "url", b.URL)
	} else if !wasAlive && alive {
		c.logger.Info("Backend is UP", "backend", b.Name, "url", b.URL)
	}
}

// Stop stops all health check goroutines.
func (c *Checker) Stop() {
	close(c.stopCh)
	c.wg.Wait()
}
