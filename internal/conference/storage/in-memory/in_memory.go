package in_memory

import (
	"context"

	"github.com/CMAK12/gonference/infra/db/in-memory/dragonfly"
)

type inMemory interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value string, opts ...dragonfly.SetOption) (bool, error)
}

type Storage struct {
	client inMemory
}

func NewStorage(client inMemory) *Storage {
	return &Storage{
		client: client,
	}
}
