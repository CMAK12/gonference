package signaling

import (
	"context"
	"errors"
	"fmt"
)

var ErrOwnerNotResolved = errors.New("room owner not resolved")

func (s *Signaling) resolveOwner(ctx context.Context, roomID string) (string, error) {
	for range s.maxResolveRetries + 1 {
		owner, err := s.inMemory.GetRoom(ctx, roomID)
		if err != nil {
			return "", fmt.Errorf("get room: %w", err)
		}
		if owner != "" {
			return owner, nil
		}

		ok, err := s.inMemory.ClaimRoom(ctx, roomID, s.self, s.ownershipTTL)
		if err != nil {
			return "", fmt.Errorf("claim room: %w", err)
		}
		if ok {
			return s.self, nil
		}
	}

	return "", fmt.Errorf("room %s: %w", roomID, ErrOwnerNotResolved)
}
