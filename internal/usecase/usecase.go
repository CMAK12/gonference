package usecase

import (
	"gonference/internal/usecase/sfu"
	"gonference/internal/usecase/signaling"
)

type UseCase struct {
	Signaling *signaling.Signaling
	SFU       *sfu.SFU
}

func NewUseCase() (*UseCase, error) {
	sfu, err := sfu.New()
	if err != nil {
		return nil, err
	}
	s := signaling.New(sfu)

	return &UseCase{
		Signaling: s,
		SFU:       sfu,
	}, nil
}
