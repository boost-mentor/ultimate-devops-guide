package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/boost-mentor/kubernetes-production-labs/apps/devops-may-cry/internal/service"
	"github.com/boost-mentor/kubernetes-production-labs/apps/devops-may-cry/internal/store"
)

var (
	version   = "dev"
	commit    = "unknown"
	buildDate = "unknown"
)

func main() {
	healthcheck := flag.Bool("healthcheck", false, "probe the local /livez endpoint")
	flag.Parse()
	if *healthcheck {
		if err := runHealthcheck(); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	orderStore, err := buildStore()
	if err != nil {
		logger.Error("configure store", "error", err)
		os.Exit(1)
	}
	defer orderStore.Close()

	hostname, _ := os.Hostname()
	api := service.New(orderStore, logger, hostname, envBool("LAB_ENDPOINTS_ENABLED"))
	server := &http.Server{
		Addr:              ":" + envOr("PORT", "8080"),
		Handler:           api.Handler(),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-stop
		logger.Info("shutdown requested")
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()
		if err := server.Shutdown(ctx); err != nil {
			logger.Error("graceful shutdown", "error", err)
		}
	}()

	logger.Info("devops may cry api started",
		"address", server.Addr,
		"pod", hostname,
		"store", orderStore.Kind(),
		"version", version,
		"commit", commit,
		"build_date", buildDate,
	)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Error("http server stopped", "error", err)
		os.Exit(1)
	}
	logger.Info("devops may cry api stopped")
}

func buildStore() (store.OrderStore, error) {
	// Storage is selected explicitly by configuration. Without DB_HOST the
	// service uses process memory: this keeps the Part 1 Kubernetes example
	// focused on cluster and traffic mechanics. Compose sets DB_HOST=postgres,
	// so /readyz also proves that PostgreSQL accepts connections and /orders
	// persists rows in the named Compose volume.
	if os.Getenv("DB_HOST") == "" {
		return store.NewMemory(), nil
	}
	port, err := strconv.Atoi(envOr("DB_PORT", "5432"))
	if err != nil {
		return nil, fmt.Errorf("DB_PORT: %w", err)
	}
	return store.NewPostgres(store.PostgresConfig{
		Host:     os.Getenv("DB_HOST"),
		Port:     port,
		User:     envOr("DB_USER", "devopsmaycry"),
		Password: os.Getenv("DB_PASSWORD"),
		Database: envOr("DB_NAME", "devopsmaycry"),
		SSLMode:  envOr("DB_SSLMODE", "disable"),
	})
}

func runHealthcheck() error {
	url := envOr("HEALTHCHECK_URL", "http://127.0.0.1:8080/livez")
	client := &http.Client{Timeout: 2 * time.Second}
	response, err := client.Get(url)
	if err != nil {
		return fmt.Errorf("healthcheck: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("healthcheck: status %d", response.StatusCode)
	}
	return nil
}

func envOr(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func envBool(key string) bool {
	value, err := strconv.ParseBool(os.Getenv(key))
	return err == nil && value
}
