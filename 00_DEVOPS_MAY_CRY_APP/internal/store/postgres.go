package store

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type Postgres struct {
	db *sql.DB
}

type PostgresConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Database string
	SSLMode  string
}

func NewPostgres(cfg PostgresConfig) (*Postgres, error) {
	if cfg.Host == "" || cfg.User == "" || cfg.Database == "" {
		return nil, fmt.Errorf("postgres config requires host, user and database")
	}
	if cfg.Port == 0 {
		cfg.Port = 5432
	}
	if cfg.SSLMode == "" {
		cfg.SSLMode = "disable"
	}

	dsn := &url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(cfg.User, cfg.Password),
		Host:   net.JoinHostPort(cfg.Host, strconv.Itoa(cfg.Port)),
		Path:   cfg.Database,
	}
	query := dsn.Query()
	query.Set("sslmode", cfg.SSLMode)
	dsn.RawQuery = query.Encode()

	db, err := sql.Open("pgx", dsn.String())
	if err != nil {
		return nil, fmt.Errorf("open postgres: %w", err)
	}
	db.SetMaxOpenConns(12)
	db.SetMaxIdleConns(4)
	db.SetConnMaxIdleTime(5 * time.Minute)
	db.SetConnMaxLifetime(30 * time.Minute)
	return &Postgres{db: db}, nil
}

func (s *Postgres) Create(ctx context.Context, target string, price int, handledBy string) (Order, error) {
	const query = `
		INSERT INTO orders (target, price_usd, status, handled_by)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at`
	order := Order{
		Target:    target,
		PriceUSD:  price,
		Status:    "принят — дежурный выехал",
		HandledBy: handledBy,
	}
	if err := s.db.QueryRowContext(ctx, query, target, price, order.Status, handledBy).
		Scan(&order.ID, &order.CreatedAt); err != nil {
		return Order{}, fmt.Errorf("insert order: %w", err)
	}
	order.Number = fmt.Sprintf("NS-%04d", order.ID)
	return order, nil
}

func (s *Postgres) Ping(ctx context.Context) error { return s.db.PingContext(ctx) }
func (*Postgres) Kind() string                     { return "postgres" }
func (s *Postgres) Close() error                   { return s.db.Close() }
