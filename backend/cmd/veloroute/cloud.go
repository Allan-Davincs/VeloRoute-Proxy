package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/allan-davincs/veloroute/internal/admin"
	"github.com/allan-davincs/veloroute/internal/metrics"
	"github.com/allan-davincs/veloroute/internal/proxy"
)

// runCloudServer binds proxy, admin API, and metrics on a single PORT (Heroku / Render).
func runCloudServer(port string, proxyHandler *proxy.Handler, adminServer *admin.Server, logger *slog.Logger) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/metrics":
			metrics.Handler().ServeHTTP(w, r)
		case strings.HasPrefix(r.URL.Path, "/api/"):
			adminServer.Handler().ServeHTTP(w, r)
		default:
			proxyHandler.ServeHTTP(w, r)
		}
	})

	srv := &http.Server{Addr: ":" + port, Handler: handler}

	go func() {
		logger.Info("cloud server listening", "addr", ":"+port, "mode", "single-port")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down VeloRoute...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("shutdown error", "error", err)
	}
	logger.Info("VeloRoute shutdown complete")
}
