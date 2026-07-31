package storage

import (
	"context"
)

// UnitOfWork exposes the storages that can take part in one transaction.
type UnitOfWork interface {
	WithinTransaction(ctx context.Context, fn func(ctx context.Context, uow UnitOfWork) error) error

	Conference() ConferenceStorage
}
