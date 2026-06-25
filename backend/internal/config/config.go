package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Load reads and parses the configuration file at the given path.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// Validate checks that required configuration fields are set.
func (c *Config) Validate() error {
	v := c.VeloRoute
	if v.ListenAddr == "" {
		return fmt.Errorf("listen_addr is required")
	}
	if v.AdminAddr == "" {
		return fmt.Errorf("admin_addr is required")
	}
	if v.MetricsAddr == "" {
		return fmt.Errorf("metrics_addr is required")
	}
	if len(v.Backends) == 0 {
		return fmt.Errorf("at least one backend is required")
	}
	for i, b := range v.Backends {
		if b.URL == "" {
			return fmt.Errorf("backend[%d]: url is required", i)
		}
		if b.Name == "" {
			return fmt.Errorf("backend[%d]: name is required", i)
		}
		if b.Weight <= 0 {
			return fmt.Errorf("backend[%d]: weight must be positive", i)
		}
	}
	return nil
}
