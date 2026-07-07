package service

import "github.com/CMAK12/gonference/internal/conference/entity"

type ConferenceService struct {
}

func NewConferenceService() *ConferenceService {
	return &ConferenceService{}
}

func (cs *ConferenceService) CreateConference(conf *entity.Conference) error {
	return nil
}

func (cs *ConferenceService) JoinConference(conf *entity.Conference) error {
	return nil
}
