package sfu

import (
	"sync"

	"github.com/CMAK12/gonference/internal/streaming/sfu/codecs"
	"github.com/pion/interceptor"
	"github.com/pion/webrtc/v4"
)

type SFU struct {
	mux   sync.RWMutex
	rooms map[string]*Room

	api *webrtc.API
}

func New() (*SFU, error) {
	mediaEngine := &webrtc.MediaEngine{}
	for _, codec := range codecs.Video {
		if err := mediaEngine.RegisterCodec(codec, webrtc.RTPCodecTypeVideo); err != nil {
			return nil, err
		}
	}

	for _, codec := range codecs.Audio {
		if err := mediaEngine.RegisterCodec(codec, webrtc.RTPCodecTypeAudio); err != nil {
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

func (s *SFU) GetRoom(id string) (*Room, bool) {
	s.mux.RLock()
	defer s.mux.RUnlock()

	room, exists := s.rooms[id]
	return room, exists
}

func (s *SFU) GetOrCreateRoom(id string) *Room {
	s.mux.Lock()
	defer s.mux.Unlock()

	room, exists := s.rooms[id]
	if !exists {
		room = NewRoom(s.api, id, func() {
			s.removeRoom(room)
		})
		s.rooms[id] = room
	}

	return room
}

func (s *SFU) removeRoom(room *Room) {
	s.mux.Lock()
	if s.rooms[room.ID()] == room {
		delete(s.rooms, room.ID())
	}
	s.mux.Unlock()

	room.Close()
}

func (s *SFU) RemoveRoom(id string) {
	s.mux.Lock()
	room, ok := s.rooms[id]
	delete(s.rooms, id)
	s.mux.Unlock()

	if ok {
		room.Close()
	}
}

func (s *SFU) Close() {
	s.mux.Lock()
	rooms := make([]*Room, 0, len(s.rooms))
	for _, room := range s.rooms {
		rooms = append(rooms, room)
	}
	clear(s.rooms)
	s.mux.Unlock()

	for _, room := range rooms {
		room.Close()
	}
}
