package sfu

import (
	"sync"

	"github.com/pion/interceptor"
	"github.com/pion/webrtc/v3"
)

type SFU struct {
	api *webrtc.API

	mux   sync.RWMutex
	rooms map[string]*Room
}

func New() (*SFU, error) {
	mediaEngine := &webrtc.MediaEngine{}
	for _, codec := range videoCodecs {
		if err := mediaEngine.RegisterCodec(codec, webrtc.RTPCodecTypeVideo); err != nil {
			return nil, err
		}
	}

	interceptorRegistry := &interceptor.Registry{}
	if err := webrtc.RegisterDefaultInterceptors(mediaEngine, interceptorRegistry); err != nil {
		return nil, err
	}

	api := webrtc.NewAPI(
		webrtc.WithMediaEngine(mediaEngine),
		webrtc.WithInterceptorRegistry(interceptorRegistry),
	)

	return &SFU{
		api:   api,
		rooms: make(map[string]*Room),
	}, nil
}

func (s *SFU) GetOrCreateRoom(id string) *Room {
	s.mux.Lock()
	defer s.mux.Unlock()

	room, ok := s.rooms[id]
	if !ok {
		room = NewRoom(s.api, id)
		s.rooms[id] = room
	}

	return room
}

func (s *SFU) RemoveRoom(id string) {
	s.mux.Lock()
	defer s.mux.Unlock()

	delete(s.rooms, id)
}

func (s *SFU) Close() {
	s.mux.Lock()
	defer s.mux.Unlock()

	for _, room := range s.rooms {
		room.Close()
	}
}
