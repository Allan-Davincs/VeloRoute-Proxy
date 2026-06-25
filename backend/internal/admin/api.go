package admin

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/allan-davincs/veloroute/internal/balancer"
	"github.com/allan-davincs/veloroute/internal/config"
	"github.com/allan-davincs/veloroute/internal/logger"
	"github.com/allan-davincs/veloroute/internal/metrics"
	"github.com/allan-davincs/veloroute/internal/proxy"
)

// Server provides the admin REST API and SSE log stream.
type Server struct {
	cfg         *config.Config
	balancer    balancer.Balancer
	metrics     *metrics.Registry
	accessLog   *logger.AccessLogger
	logger      *slog.Logger
	algorithm   string
	mu          sync.RWMutex
	totalReqs   int64
	lastReqTime time.Time
	reqCount    int64
}

// NewServer creates a new admin API server.
func NewServer(cfg *config.Config, b balancer.Balancer, m *metrics.Registry, log *logger.AccessLogger, logger *slog.Logger, algorithm string) *Server {
	return &Server{
		cfg:       cfg,
		balancer:  b,
		metrics:   m,
		accessLog: log,
		logger:    logger,
		algorithm: algorithm,
	}
}

// Handler returns the HTTP handler for the admin API.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/backends", s.handleBackends)
	mux.HandleFunc("/api/backends/", s.handleBackendByURL)
	mux.HandleFunc("/api/config", s.handleConfig)
	mux.HandleFunc("/api/config/algorithm", s.handleAlgorithm)
	mux.HandleFunc("/api/metrics", s.handleMetrics)
	mux.HandleFunc("/api/logs/stream", s.handleLogStream)
	return corsMiddleware(mux)
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Server) handleBackends(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.writeJSON(w, s.backendList())
	case http.MethodPost:
		s.addBackend(w, r)
	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleBackendByURL(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/backends/")
	parts := strings.Split(path, "/")
	if len(parts) == 0 || parts[0] == "" {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	encodedURL := parts[0]
	rawURL, err := base64.URLEncoding.DecodeString(encodedURL)
	if err != nil {
		http.Error(w, "Invalid URL encoding", http.StatusBadRequest)
		return
	}
	backendURL := string(rawURL)

	if len(parts) == 2 && parts[1] == "weight" && r.Method == http.MethodPut {
		s.updateWeight(w, r, backendURL)
		return
	}

	if r.Method == http.MethodDelete {
		s.balancer.RemoveBackend(backendURL)
		s.writeJSON(w, map[string]string{"status": "removed"})
		return
	}

	http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
}

func (s *Server) addBackend(w http.ResponseWriter, r *http.Request) {
	var req struct {
		URL    string `json:"url"`
		Name   string `json:"name"`
		Weight int    `json:"weight"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	if req.Weight <= 0 {
		req.Weight = 1
	}
	if err := proxy.ValidateBackendURL(req.URL); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	b := &balancer.Backend{URL: req.URL, Name: req.Name, Weight: req.Weight}
	s.balancer.AddBackend(b)
	s.writeJSON(w, map[string]string{"status": "added"})
}

func (s *Server) updateWeight(w http.ResponseWriter, r *http.Request, backendURL string) {
	var req struct {
		Weight int `json:"weight"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	for _, b := range s.balancer.GetBackends() {
		if b.URL == backendURL {
			b.Weight = req.Weight
			s.writeJSON(w, map[string]string{"status": "updated"})
			return
		}
	}
	http.Error(w, "Backend not found", http.StatusNotFound)
}

func (s *Server) handleConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	s.mu.RLock()
	algo := s.algorithm
	s.mu.RUnlock()
	s.writeJSON(w, map[string]interface{}{
		"listen_addr":  s.cfg.VeloRoute.ListenAddr,
		"admin_addr":   s.cfg.VeloRoute.AdminAddr,
		"metrics_addr": s.cfg.VeloRoute.MetricsAddr,
		"algorithm":    algo,
		"rate_limit":   s.cfg.VeloRoute.RateLimit,
		"health_check": s.cfg.VeloRoute.HealthCheck,
	})
}

func (s *Server) handleAlgorithm(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		Algorithm string `json:"algorithm"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	s.mu.Lock()
	s.algorithm = req.Algorithm
	s.mu.Unlock()
	s.writeJSON(w, map[string]string{"algorithm": req.Algorithm})
}

func (s *Server) handleMetrics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	s.writeJSON(w, s.metricsSnapshot())
}

func (s *Server) metricsSnapshot() map[string]interface{} {
	backends := s.balancer.GetBackends()
	var totalReqs int64
	var totalErrors int64
	var activeConns int64
	backendList := make([]map[string]interface{}, 0, len(backends))

	for _, b := range backends {
		reqs := atomic.LoadInt64(&b.TotalRequests)
		errs := atomic.LoadInt64(&b.TotalErrors)
		conns := atomic.LoadInt64(&b.ActiveConns)
		totalReqs += reqs
		totalErrors += errs
		activeConns += conns

		lastCheck := ""
		if t, ok := b.LastCheck.Load().(time.Time); ok {
			lastCheck = t.UTC().Format(time.RFC3339)
		}

		s.metrics.SetBackendAlive(b.URL, b.Name, b.Alive())
		s.metrics.SetActiveConnections(b.Name, conns)

		backendList = append(backendList, map[string]interface{}{
			"name":               b.Name,
			"url":                b.URL,
			"alive":              b.Alive(),
			"weight":             b.Weight,
			"active_connections": conns,
			"total_requests":     reqs,
			"total_errors":       errs,
			"last_health_check":  lastCheck,
		})
	}

	p50, p95, p99 := s.metrics.Percentiles()
	errorRate := 0.0
	if totalReqs > 0 {
		errorRate = float64(totalErrors) / float64(totalReqs) * 100
	}

	rps := s.requestsPerSecond(totalReqs)

	return map[string]interface{}{
		"total_requests":      totalReqs,
		"requests_per_second": rps,
		"error_rate_percent":  errorRate,
		"active_connections":  activeConns,
		"p50_latency_ms":      p50,
		"p95_latency_ms":      p95,
		"p99_latency_ms":      p99,
		"uptime_seconds":      s.metrics.UptimeSeconds(),
		"backends":            backendList,
	}
}

func (s *Server) requestsPerSecond(totalReqs int64) float64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now()
	elapsed := now.Sub(s.lastReqTime).Seconds()
	if elapsed < 1 {
		elapsed = 1
	}
	delta := float64(totalReqs - s.reqCount)
	s.reqCount = totalReqs
	s.lastReqTime = now
	return delta / elapsed
}

func (s *Server) backendList() []map[string]interface{} {
	backends := s.balancer.GetBackends()
	result := make([]map[string]interface{}, 0, len(backends))
	for _, b := range backends {
		lastCheck := ""
		if t, ok := b.LastCheck.Load().(time.Time); ok {
			lastCheck = t.UTC().Format(time.RFC3339)
		}
		result = append(result, map[string]interface{}{
			"name":               b.Name,
			"url":                b.URL,
			"alive":              b.Alive(),
			"weight":             b.Weight,
			"active_connections": atomic.LoadInt64(&b.ActiveConns),
			"total_requests":     atomic.LoadInt64(&b.TotalRequests),
			"total_errors":       atomic.LoadInt64(&b.TotalErrors),
			"last_health_check":  lastCheck,
		})
	}
	return result
}

func (s *Server) handleLogStream(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming not supported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	ch := s.accessLog.Subscribe()
	defer s.accessLog.Unsubscribe(ch)

	notify := r.Context().Done()
	for {
		select {
		case entry, open := <-ch:
			if !open {
				return
			}
			data, err := logger.MarshalEntry(entry)
			if err != nil {
				continue
			}
			fmt.Fprintf(w, "data: %s\n\n", data)
			flusher.Flush()
		case <-notify:
			return
		}
	}
}

func (s *Server) writeJSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(v); err != nil {
		s.logger.Error("failed to encode JSON", "error", err)
	}
}

// Algorithm returns the current load balancing algorithm.
func (s *Server) Algorithm() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.algorithm
}

// SetAlgorithm updates the current algorithm name.
func (s *Server) SetAlgorithm(algo string) {
	s.mu.Lock()
	s.algorithm = algo
	s.mu.Unlock()
}

// DrainBody reads and closes a request body.
func DrainBody(r *http.Request) {
	io.Copy(io.Discard, r.Body)
	r.Body.Close()
}
