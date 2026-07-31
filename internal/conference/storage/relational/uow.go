package relational

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5"

	"github.com/CMAK12/gonference/infra/db/relational/postgres"
	"github.com/CMAK12/gonference/internal/conference/storage"
)

type UoW struct {
	client  *postgres.Client
	querier Querier

	conference *Conference
}

func NewUnitOfWork(client *postgres.Client) storage.UnitOfWork {
	return &UoW{
		client:     client,
		querier:    client,
		conference: NewConferenceStorage(client),
	}
}

func (s *UoW) Conference() storage.ConferenceStorage {
	if s.conference == nil {
		s.conference = NewConferenceStorage(s.client)
	}

	return s.conference
}

func (s *UoW) WithinTransaction(ctx context.Context, fn func(ctx context.Context, uow storage.UnitOfWork) error) error {
	tx, err := s.begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			if err := tx.Rollback(ctx); err != nil {
				slog.ErrorContext(ctx, "Failed to roll back panicking transaction", slog.String("error", err.Error()))
			}

			panic(p)
		}
	}()

	if err := fn(ctx, s.withQuerier(tx)); err != nil {
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			return fmt.Errorf("%w (rollback transaction: %v)", err, rbErr)
		}

		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

func (s *UoW) begin(ctx context.Context) (pgx.Tx, error) {
	if tx, ok := s.querier.(pgx.Tx); ok {
		return tx.Begin(ctx)
	}

	return s.client.Begin(ctx)
}

func (s *UoW) withQuerier(querier Querier) *UoW {
	return &UoW{
		client:     s.client,
		querier:    querier,
		conference: NewConferenceStorage(querier),
	}
}
