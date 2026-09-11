package store

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"
)

type Memory struct {
	next atomic.Int64
}

func NewMemory() *Memory { return &Memory{} }

func (s *Memory) Create(_ context.Context, target string, price int, handledBy string) (Order, error) {
	id := s.next.Add(1)
	return Order{
		ID:        id,
		Number:    fmt.Sprintf("NS-%04d", id),
		Target:    target,
		PriceUSD:  price,
		Status:    "принят — дежурный выехал",
		HandledBy: handledBy,
		CreatedAt: time.Now().UTC(),
	}, nil
}

func (*Memory) Ping(context.Context) error { return nil }
func (*Memory) Kind() string               { return "memory" }
func (*Memory) Close() error               { return nil }
