package sfu

import (
	"sync"

	"github.com/pion/interceptor"
	"github.com/pion/webrtc/v3"
)

type Signaling interface {
	WriteMessage(messageType int, data []byte) error
}

type SFU struct {
	api   *webrtc.API
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

	sfu := &SFU{
		api:   api,
		rooms: make(map[string]*Room),
	}

	return sfu, nil
}

func (u *SFU) GetPeer(roomID, peerID string) *Peer {
	u.mux.RLock()
	defer u.mux.RUnlock()

	if room, ok := u.rooms[roomID]; ok {
		room.mux.RLock()
		defer room.mux.RUnlock()
		peer := room.peers[peerID]
		return peer
	}

	return nil
}

func (u *SFU) AddPeer(ws Signaling, roomID, peerID string) (*Peer, error) {
	u.mux.Lock()
	defer u.mux.Unlock()

	var room *Room
	if r, ok := u.rooms[roomID]; ok {
		room = r
	} else {
		room = NewRoom(roomID)
		u.rooms[roomID] = room
	}

	return room.AddPeer(u.api, ws, peerID)
}

func forwardRTP(remote *webrtc.TrackRemote, local *webrtc.TrackLocalStaticRTP) {
	buf := make([]byte, 1500)

	for {
		n, _, err := remote.Read(buf)
		if err != nil {
			return
		}

		if _, err := local.Write(buf[:n]); err != nil {
			return
		}
	}
}
