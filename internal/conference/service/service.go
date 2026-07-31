package service

import (
	"github.com/CMAK12/gonference/internal/conference/storage"
)

type (
	InMemoryStorage interface {
	}
)

type Service struct {
	inMemory InMemoryStorage
	uow      storage.UnitOfWork
}

func NewService(inMemory InMemoryStorage, uow storage.UnitOfWork) *Service {
	return &Service{
		inMemory: inMemory,
		uow:      uow,
	}
}
