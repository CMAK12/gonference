package usecase

import (
	"gonference/internal/usecase/conference"
	"gonference/internal/usecase/sfu"
	"gonference/internal/usecase/signaling"
)

type UseCase struct {
	Signaling  *signaling.Signaling
	SFU        *sfu.SFU
	Conference *conference.Registry
}

func NewUseCase() (*UseCase, error) {
	sfu, err := sfu.New()
	if err != nil {
		return nil, err
	}
	s := signaling.New(sfu)

	return &UseCase{
		Signaling:  s,
		SFU:        sfu,
		Conference: conference.New(),
	}, nil
}
