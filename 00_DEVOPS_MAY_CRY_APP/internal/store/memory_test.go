package store

import (
	"context"
	"testing"
)

func TestMemoryOrderNumbersAreMonotonic(t *testing.T) {
	s := NewMemory()
	first, err := s.Create(context.Background(), "disk-full", 800, "pod-a")
	if err != nil {
		t.Fatal(err)
	}
	second, err := s.Create(context.Background(), "dns-ghost", 3000, "pod-b")
	if err != nil {
		t.Fatal(err)
	}
	if first.Number != "NS-0001" || second.Number != "NS-0002" {
		t.Fatalf("numbers = %q, %q", first.Number, second.Number)
	}
}
