package main

import (
	"context"
	"flag"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/allan-davincs/veloroute/internal/admin"
	"github.com/allan-davincs/veloroute/internal/balancer"
	"github.com/allan-davincs/veloroute/internal/config"
	"github.com/allan-davincs/veloroute/internal/health"
	"github.com/allan-davincs/veloroute/internal/logger"
	"github.com/allan-davincs/veloroute/internal/metrics"
	"github.com/allan-davincs/veloroute/internal/proxy"
	"github.com/allan-davincs/veloroute/internal/ratelimit"
)

func main() {
	configPath := flag.String("config", "./config.yaml", "path to config file")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	appLogger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(appLogger)

	metricsReg := metrics.NewRegistry()

	rl := ratelimit.New(
		cfg.VeloRoute.RateLimit.Enabled,
		cfg.VeloRoute.RateLimit.RequestsPerSecond,
		cfg.VeloRoute.RateLimit.Burst,
	)

	pool, err := balancer.NewPool(cfg.VeloRoute.LoadBalancing.Algorithm)
	if err != nil {
		appLogger.Error("failed to create balancer pool", "error", err)
		os.Exit(1)
	}

	for _, bc := range cfg.VeloRoute.Backends {
		pool.AddBackend(&balancer.Backend{
			URL:    bc.URL,
			Name:   bc.Name,
			Weight: bc.Weight,
		})
	}

	accessLog := logger.NewAccessLogger()

	healthChecker := health.NewChecker(
		cfg.VeloRoute.HealthCheck.Enabled,
		cfg.VeloRoute.HealthCheck.IntervalSeconds,
		cfg.VeloRoute.HealthCheck.TimeoutSeconds,
		cfg.VeloRoute.HealthCheck.Path,
		pool,
		appLogger,
	)
	healthChecker.Start()

	adminServer := admin.NewServer(cfg, pool, metricsReg, accessLog, appLogger)
	proxyHandler := proxy.NewHandler(pool, rl, metricsReg, accessLog)

	proxySrv := &http.Server{Addr: cfg.VeloRoute.ListenAddr, Handler: proxyHandler}
	adminSrv := &http.Server{Addr: cfg.VeloRoute.AdminAddr, Handler: adminServer.Handler()}
	metricsSrv := &http.Server{Addr: cfg.VeloRoute.MetricsAddr, Handler: metrics.Handler()}

	go func() {
		appLogger.Info("proxy server listening", "addr", cfg.VeloRoute.ListenAddr)
		if err := proxySrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			appLogger.Error("proxy server error", "error", err)
			os.Exit(1)
		}
	}()

	go func() {
		appLogger.Info("admin API listening", "addr", cfg.VeloRoute.AdminAddr)
		if err := adminSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			appLogger.Error("admin server error", "error", err)
			os.Exit(1)
		}
	}()

	go func() {
		appLogger.Info("metrics server listening", "addr", cfg.VeloRoute.MetricsAddr)
		if err := metricsSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			appLogger.Error("metrics server error", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	appLogger.Info("shutting down VeloRoute...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	healthChecker.Stop()
	rl.Stop()

	for _, srv := range []*http.Server{proxySrv, adminSrv, metricsSrv} {
		if err := srv.Shutdown(ctx); err != nil {
			appLogger.Error("server shutdown error", "error", err)
		}
	}

	appLogger.Info("VeloRoute shutdown complete")
}
