package store

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func TestPostgresStoreIntegration(t *testing.T) {
	if os.Getenv("DEVOPS_MAY_CRY_POSTGRES_TEST") != "1" {
		t.Skip("set DEVOPS_MAY_CRY_POSTGRES_TEST=1 for the disposable PostgreSQL integration test")
	}

	port, err := strconv.Atoi(envOr("TEST_DB_PORT", "5432"))
	if err != nil {
		t.Fatalf("TEST_DB_PORT: %v", err)
	}
	config := PostgresConfig{
		Host:     envOr("TEST_DB_HOST", "127.0.0.1"),
		Port:     port,
		User:     envOr("TEST_DB_USER", "devopsmaycry"),
		Password: os.Getenv("TEST_DB_PASSWORD"),
		Database: envOr("TEST_DB_NAME", "devopsmaycry"),
		SSLMode:  "disable",
	}
	if config.Password == "" {
		t.Fatal("TEST_DB_PASSWORD is required for the disposable integration database")
	}

	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		config.Host, config.Port, config.User, config.Password, config.Database, config.SSLMode,
	)
	database, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()

	migrationPath := filepath.Join("..", "..", "migrations", "001_init.sql")
	migration, err := os.ReadFile(migrationPath)
	if err != nil {
		t.Fatalf("read migration: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if _, err := database.ExecContext(ctx, string(migration)); err != nil {
		t.Fatalf("apply migration: %v", err)
	}

	orderStore, err := NewPostgres(config)
	if err != nil {
		t.Fatal(err)
	}
	defer orderStore.Close()
	if err := orderStore.Ping(ctx); err != nil {
		t.Fatalf("ping: %v", err)
	}

	target := fmt.Sprintf("integration-%d", time.Now().UnixNano())
	t.Cleanup(func() {
		_, _ = database.ExecContext(context.Background(), "DELETE FROM orders WHERE target = $1", target)
	})
	created, err := orderStore.Create(ctx, target, 4242, "ci-pod")
	if err != nil {
		t.Fatalf("create order: %v", err)
	}
	if created.Target != target || created.PriceUSD != 4242 || created.HandledBy != "ci-pod" {
		t.Fatalf("unexpected order: %#v", created)
	}
}

func envOr(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
