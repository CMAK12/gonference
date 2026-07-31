package in_memory

import (
	"context"
	"time"
)

type inMemory interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value string, expiration time.Duration) error
}

type Storage struct {
	client inMemory
}

func NewStorage(client inMemory) *Storage {
	return &Storage{
		client: client,
	}
}
