package config

// Config holds the full VeloRoute configuration.
type Config struct {
	VeloRoute VeloRouteConfig `yaml:"veloroute"`
}

// VeloRouteConfig contains top-level VeloRoute settings.
type VeloRouteConfig struct {
	ListenAddr  string              `yaml:"listen_addr"`
	AdminAddr   string              `yaml:"admin_addr"`
	MetricsAddr string              `yaml:"metrics_addr"`
	LoadBalancing LoadBalancingConfig `yaml:"load_balancing"`
	HealthCheck   HealthCheckConfig   `yaml:"health_check"`
	RateLimit     RateLimitConfig     `yaml:"rate_limit"`
	Backends      []BackendConfig     `yaml:"backends"`
}

// LoadBalancingConfig configures the load balancing algorithm.
type LoadBalancingConfig struct {
	Algorithm string `yaml:"algorithm"`
}

// HealthCheckConfig configures active backend health checks.
type HealthCheckConfig struct {
	Enabled          bool   `yaml:"enabled"`
	IntervalSeconds  int    `yaml:"interval_seconds"`
	TimeoutSeconds   int    `yaml:"timeout_seconds"`
	Path             string `yaml:"path"`
}

// RateLimitConfig configures per-IP token bucket rate limiting.
type RateLimitConfig struct {
	Enabled           bool `yaml:"enabled"`
	RequestsPerSecond int  `yaml:"requests_per_second"`
	Burst             int  `yaml:"burst"`
}

// BackendConfig defines a single upstream backend server.
type BackendConfig struct {
	URL    string `yaml:"url"`
	Weight int    `yaml:"weight"`
	Name   string `yaml:"name"`
}
