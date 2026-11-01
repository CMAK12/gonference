package signaling

import (
	"context"
	"time"

	gnpb "github.com/CMAK12/gonference/internal/gen/streaming/v1"
	"github.com/CMAK12/gonference/internal/streaming/sfu"

	"golang.org/x/sync/singleflight"
)

type (
	SFU interface {
		GetRoom(id string) (*sfu.Room, bool)
		GetOrCreateRoom(id string) *sfu.Room
	}

	InMemory interface {
		GetRoom(ctx context.Context, roomID string) (string, error)
		ClaimRoom(ctx context.Context, roomID, owner string, ttl time.Duration) (bool, error)
		RefreshRoom(ctx context.Context, roomID, owner string, ttl time.Duration) (bool, error)
		RemoveRoom(ctx context.Context, roomID string) error
	}

	RelayClient interface {
		ConnectRelay(ctx context.Context, podAddr string, req *gnpb.ConnectRelayRequest) (*gnpb.ConnectRelayResponse, error)
		DisconnectRelay(ctx context.Context, podAddr string, req *gnpb.DisconnectRelayRequest) error
	}
)

type Signaling struct {
	self              string
	ownershipTTL      time.Duration
	maxResolveRetries int
	sfu               SFU
	inMemory          InMemory
	relay             RelayClient

	relayGroup singleflight.Group
}

func New(self string, ownershipTTL time.Duration, maxResolveRetries int, sfu SFU, inMemory InMemory, relay RelayClient) *Signaling {
	return &Signaling{
		self:              self,
		ownershipTTL:      ownershipTTL,
		maxResolveRetries: maxResolveRetries,
		sfu:               sfu,
		inMemory:          inMemory,
		relay:             relay,
	}
}
