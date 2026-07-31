package service

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/CMAK12/gonference/internal/conference/entity"
	"github.com/CMAK12/gonference/internal/conference/storage"
)

func (s *Service) CreateConference(ctx context.Context, conf *entity.Conference) error {
	if conf.ID == "" {
		conf.ID = uuid.NewString()
	}
	if conf.Token == "" {
		conf.Token = uuid.NewString()
	}
	if conf.CreatedAt.IsZero() {
		conf.CreatedAt = time.Now().UTC()
	}

	return s.uow.WithinTransaction(ctx, func(ctx context.Context, uow storage.UnitOfWork) error {
		return uow.Conference().CreateConference(ctx, conf)
	})
}

func (s *Service) JoinConference(ctx context.Context, conf *entity.Conference) error {
	return s.uow.Conference().JoinConference(ctx, conf)
}
