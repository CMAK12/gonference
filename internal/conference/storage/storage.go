package storage

import (
	"context"
	"errors"

	"github.com/CMAK12/gonference/internal/conference/entity"
)

// ErrNotImplemented is returned by storage methods that have no implementation yet.
var ErrNotImplemented = errors.New("not implemented")

// ErrConferenceNotFound is returned when no conference exists for the requested identifier.
var ErrConferenceNotFound = errors.New("conference not found")

// ConferenceStorage persists conferences and their membership.
type ConferenceStorage interface {
	CreateConference(ctx context.Context, conf *entity.Conference) error
	JoinConference(ctx context.Context, conf *entity.Conference) error
}
