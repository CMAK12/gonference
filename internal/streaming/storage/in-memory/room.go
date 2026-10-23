package in_memory

import (
	"context"
	"time"

	"github.com/CMAK12/gonference/infra/db/in-memory/dragonfly"
)

func roomKey(roomID string) string { return "conference:room:" + roomID }

func (s *Storage) GetRoom(ctx context.Context, roomID string) (string, error) {
	return s.client.Get(ctx, roomKey(roomID))
}

func (s *Storage) ClaimRoom(ctx context.Context, roomID, owner string, ttl time.Duration) (bool, error) {
	return s.client.Set(ctx, roomKey(roomID), owner, dragonfly.NX(), dragonfly.WithTTL(ttl))
}

func (s *Storage) RefreshRoom(ctx context.Context, roomID, owner string, ttl time.Duration) (bool, error) {
	return s.client.Set(ctx, roomKey(roomID), owner, dragonfly.XX(), dragonfly.WithTTL(ttl))
}

func (s *Storage) RemoveRoom(ctx context.Context, roomID string) error {
	return s.client.Del(ctx, roomKey(roomID))
}
