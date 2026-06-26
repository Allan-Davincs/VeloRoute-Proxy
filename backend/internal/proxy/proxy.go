package proxy

import (
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"sync/atomic"
	"time"

	"github.com/allan-davincs/veloroute/internal/balancer"
	"github.com/allan-davincs/veloroute/internal/logger"
	"github.com/allan-davincs/veloroute/internal/metrics"
	"github.com/allan-davincs/veloroute/internal/ratelimit"
	"github.com/google/uuid"
)

type responseWriter struct {
	http.ResponseWriter
	status int
	bytes  int64
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	if rw.status == 0 {
		rw.status = http.StatusOK
	}
	n, err := rw.ResponseWriter.Write(b)
	rw.bytes += int64(n)
	return n, err
}

// Handler proxies HTTP requests to backend servers.
type Handler struct {
	pool      *balancer.Pool
	limiter   *ratelimit.Limiter
	metrics   *metrics.Registry
	accessLog *logger.AccessLogger
}

// NewHandler creates a new reverse proxy handler.
func NewHandler(pool *balancer.Pool, limiter *ratelimit.Limiter, m *metrics.Registry, log *logger.AccessLogger) *Handler {
	return &Handler{
		pool:      pool,
		limiter:   limiter,
		metrics:   m,
		accessLog: log,
	}
}

// ServeHTTP handles an incoming HTTP request.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	clientIP := clientIPFromRequest(r)

	if !h.limiter.Allow(clientIP) {
		retryAfter := h.limiter.RetryAfterSeconds(clientIP)
		h.metrics.RecordError("none", "rate_limited")
		ratelimit.WriteRateLimitResponse(w, retryAfter)
		return
	}

	backend, err := h.pool.Balancer().Next(clientIP)
	if err != nil {
		h.metrics.RecordError("none", "no_backend")
		http.Error(w, "Service Unavailable", http.StatusServiceUnavailable)
		return
	}

	leastConn := h.pool.Algorithm() == "least_connections"
	if leastConn {
		atomic.AddInt64(&backend.ActiveConns, 1)
		defer atomic.AddInt64(&backend.ActiveConns, -1)
		h.metrics.SetActiveConnections(backend.Name, atomic.LoadInt64(&backend.ActiveConns))
	}

	atomic.AddInt64(&backend.TotalRequests, 1)

	target, err := url.Parse(backend.URL)
	if err != nil {
		atomic.AddInt64(&backend.TotalErrors, 1)
		h.metrics.RecordError(backend.Name, "invalid_url")
		http.Error(w, "Bad Gateway", http.StatusBadGateway)
		return
	}

	proxy := httputil.NewSingleHostReverseProxy(target)
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		req.Header.Set("X-Forwarded-For", appendXFF(req.Header.Get("X-Forwarded-For"), clientIP))
		req.Header.Set("X-Forwarded-Host", r.Host)
		req.Header.Set("X-Real-IP", clientIP)
		req.Header.Set("X-VeloRoute-Backend", backend.Name)
	}

	rw := &responseWriter{ResponseWriter: w, status: http.StatusOK}
	requestID := uuid.New().String()

	proxy.ErrorHandler = func(rw http.ResponseWriter, _ *http.Request, _ error) {
		atomic.AddInt64(&backend.TotalErrors, 1)
		h.metrics.RecordError(backend.Name, "proxy_error")
		http.Error(rw, "Bad Gateway", http.StatusBadGateway)
	}

	proxy.ServeHTTP(rw, r)

	duration := time.Since(start)
	h.metrics.RecordRequest(backend.Name, r.Method, rw.status, duration)

	h.accessLog.Log(logger.AccessEntry{
		ClientIP:   clientIP,
		Method:     r.Method,
		Path:       r.URL.Path,
		Query:      r.URL.RawQuery,
		Status:     rw.status,
		DurationMS: float64(duration.Milliseconds()),
		BytesSent:  rw.bytes,
		Backend:    backend.Name,
		BackendURL: backend.URL,
		UserAgent:  r.UserAgent(),
		RequestID:  requestID,
	})
}

func clientIPFromRequest(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		return strings.TrimSpace(parts[0])
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

func appendXFF(existing, clientIP string) string {
	if existing == "" {
		return clientIP
	}
	return existing + ", " + clientIP
}

// BuildBackends creates Backend structs from config entries.
func BuildBackends(cfgs []struct {
	URL    string
	Weight int
	Name   string
}) []*balancer.Backend {
	var backends []*balancer.Backend
	for _, c := range cfgs {
		backends = append(backends, &balancer.Backend{
			URL:    c.URL,
			Name:   c.Name,
			Weight: c.Weight,
		})
	}
	return backends
}

// RegisterBackends adds backends from config to the balancer.
func RegisterBackends(b balancer.Balancer, backends []*balancer.Backend) {
	for _, backend := range backends {
		b.AddBackend(backend)
	}
}

// ValidateBackendURL checks that a backend URL is valid.
func ValidateBackendURL(raw string) error {
	u, err := url.Parse(raw)
	if err != nil {
		return err
	}
	if u.Scheme == "" || u.Host == "" {
		return fmt.Errorf("invalid backend URL: %s", raw)
	}
	return nil
}
