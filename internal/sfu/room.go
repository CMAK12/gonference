package sfu

import (
	"sync"

	"github.com/pion/webrtc/v3"
)

type Room struct {
	id     string
	mux    sync.RWMutex
	peers  map[string]*Peer
	tracks map[string]*webrtc.TrackRemote
}

func NewRoom(id string) *Room {
	return &Room{
		id:     id,
		peers:  make(map[string]*Peer),
		tracks: make(map[string]*webrtc.TrackRemote),
	}
}

func (r *Room) AddPeer(api *webrtc.API, ws Signaling, peerID string) (*Peer, error) {
	r.mux.Lock()
	defer r.mux.Unlock()

	peer, err := NewPeer(api, ws, peerID, r)
	if err != nil {
		return nil, err
	}

	r.peers[peerID] = peer

	for _, track := range r.tracks {
		local, err := webrtc.NewTrackLocalStaticRTP(
			track.Codec().RTPCodecCapability,
			track.ID(),
			track.StreamID(),
		)
		if err != nil {
			return nil, err
		}

		_, err = peer.conn.AddTrack(local)
		if err != nil {
			return nil, err
		}

		peer.outTracks[track.ID()] = local

		go forwardRTP(track, local)
	}

	return peer, nil
}

func (r *Room) AddIncomingTrack(from *Peer, remote *webrtc.TrackRemote) {
	r.mux.Lock()
	defer r.mux.Unlock()

	r.tracks[remote.ID()] = remote

	for _, peer := range r.peers {
		// do not send to self
		if peer == from {
			continue
		}

		local, err := webrtc.NewTrackLocalStaticRTP(
			remote.Codec().RTPCodecCapability,
			remote.ID(),
			remote.StreamID(),
		)
		if err != nil {
			continue
		}

		_, err = peer.conn.AddTrack(local)
		if err != nil {
			continue
		}

		peer.outTracks[remote.ID()] = local

		if err := peer.Renegotiate(); err != nil {
			continue
		}

		from.SendPLI(uint32(remote.SSRC()))

		go forwardRTP(remote, local)
	}
}
