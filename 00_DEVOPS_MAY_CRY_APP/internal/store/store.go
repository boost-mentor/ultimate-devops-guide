package store

import (
	"context"
	"time"
)

// Order is the durable business object returned by the dispatch API.
type Order struct {
	ID        int64     `json:"-"`
	Number    string    `json:"order"`
	Target    string    `json:"target"`
	PriceUSD  int       `json:"price_usd"`
	Status    string    `json:"status"`
	HandledBy string    `json:"handled_by"`
	CreatedAt time.Time `json:"-"`
}

// OrderStore keeps HTTP and persistence concerns separate. The API can run
// in memory for network/resource labs and use PostgreSQL in the full stack.
type OrderStore interface {
	Create(context.Context, string, int, string) (Order, error)
	Ping(context.Context) error
	Kind() string
	Close() error
}
