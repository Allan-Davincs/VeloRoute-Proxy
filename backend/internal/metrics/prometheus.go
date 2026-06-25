package metrics

import (
	"net/http"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var durationBuckets = []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.5, 5.0}

// Registry holds all VeloRoute Prometheus metrics.
type Registry struct {
	RequestsTotal    *prometheus.CounterVec
	ErrorsTotal      *prometheus.CounterVec
	RequestDuration  *prometheus.HistogramVec
	ActiveConns      *prometheus.GaugeVec
	BackendAlive     *prometheus.GaugeVec
	DurationSummary  *prometheus.SummaryVec
	startTime        time.Time
	mu               sync.Mutex
	latencySamples   []float64
	maxSamples       int
}

// NewRegistry creates and registers all VeloRoute metrics.
func NewRegistry() *Registry {
	r := &Registry{
		startTime:      time.Now(),
		maxSamples:     10000,
		latencySamples: make([]float64, 0, 10000),
		RequestsTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "veloroute_requests_total",
			Help: "Total number of requests proxied",
		}, []string{"backend", "method", "status_code"}),
		ErrorsTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "veloroute_errors_total",
			Help: "Total number of proxy errors",
		}, []string{"backend", "error_type"}),
		RequestDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "veloroute_request_duration_seconds",
			Help:    "Request duration in seconds",
			Buckets: durationBuckets,
		}, []string{"backend", "method"}),
		ActiveConns: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "veloroute_active_connections",
			Help: "Active connections per backend",
		}, []string{"backend"}),
		BackendAlive: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "veloroute_backend_alive",
			Help: "Backend alive status (1=alive, 0=dead)",
		}, []string{"backend", "name"}),
		DurationSummary: prometheus.NewSummaryVec(prometheus.SummaryOpts{
			Name:       "veloroute_request_duration_summary",
			Help:       "Request duration summary",
			Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.95: 0.01, 0.99: 0.001},
		}, []string{"backend"}),
	}

	prometheus.MustRegister(
		r.RequestsTotal,
		r.ErrorsTotal,
		r.RequestDuration,
		r.ActiveConns,
		r.BackendAlive,
		r.DurationSummary,
	)
	return r
}

// RecordRequest records metrics for a completed request.
func (r *Registry) RecordRequest(backend, method string, statusCode int, duration time.Duration) {
	status := http.StatusText(statusCode)
	if status == "" {
		status = "unknown"
	}
	r.RequestsTotal.WithLabelValues(backend, method, itoa(statusCode)).Inc()
	secs := duration.Seconds()
	r.RequestDuration.WithLabelValues(backend, method).Observe(secs)
	r.DurationSummary.WithLabelValues(backend).Observe(secs)

	r.mu.Lock()
	if len(r.latencySamples) >= r.maxSamples {
		r.latencySamples = r.latencySamples[1:]
	}
	r.latencySamples = append(r.latencySamples, duration.Seconds()*1000)
	r.mu.Unlock()
}

// RecordError increments the error counter.
func (r *Registry) RecordError(backend, errorType string) {
	r.ErrorsTotal.WithLabelValues(backend, errorType).Inc()
}

// SetActiveConnections sets the active connection gauge for a backend.
func (r *Registry) SetActiveConnections(backend string, conns int64) {
	r.ActiveConns.WithLabelValues(backend).Set(float64(conns))
}

// SetBackendAlive sets the backend alive gauge.
func (r *Registry) SetBackendAlive(backend, name string, alive bool) {
	val := 0.0
	if alive {
		val = 1.0
	}
	r.BackendAlive.WithLabelValues(backend, name).Set(val)
}

// UptimeSeconds returns seconds since registry creation.
func (r *Registry) UptimeSeconds() float64 {
	return time.Since(r.startTime).Seconds()
}

// Percentiles returns p50, p95, p99 latency in milliseconds.
func (r *Registry) Percentiles() (p50, p95, p99 float64) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if len(r.latencySamples) == 0 {
		return 0, 0, 0
	}
	sorted := make([]float64, len(r.latencySamples))
	copy(sorted, r.latencySamples)
	sortFloat64s(sorted)
	p50 = percentile(sorted, 0.50)
	p95 = percentile(sorted, 0.95)
	p99 = percentile(sorted, 0.99)
	return
}

func percentile(sorted []float64, p float64) float64 {
	if len(sorted) == 0 {
		return 0
	}
	idx := int(float64(len(sorted)-1) * p)
	return sorted[idx]
}

func sortFloat64s(a []float64) {
	for i := 1; i < len(a); i++ {
		for j := i; j > 0 && a[j] < a[j-1]; j-- {
			a[j], a[j-1] = a[j-1], a[j]
		}
	}
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[i:])
}

// Handler returns the Prometheus HTTP handler.
func Handler() http.Handler {
	return promhttp.Handler()
}
