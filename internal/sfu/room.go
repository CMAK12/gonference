package sfu

import (
	"sync"

	"github.com/pion/webrtc/v3"
)

type Room struct {
	id  string
	api *webrtc.API

	mux        sync.RWMutex
	peers      map[string]*Peer
	forwarders map[string]*TrackForwarder
}

func NewRoom(api *webrtc.API, id string) *Room {
	return &Room{
		id:         id,
		api:        api,
		peers:      make(map[string]*Peer),
		forwarders: make(map[string]*TrackForwarder),
	}
}

func (r *Room) ID() string {
	return r.id
}

func (r *Room) GetPeer(id string) (*Peer, bool) {
	r.mux.RLock()
	defer r.mux.RUnlock()

	peer, ok := r.peers[id]
	return peer, ok
}

func (r *Room) AddPeer(signal Signaling, id string) (*Peer, error) {
	peer, err := NewPeer(r.api, signal, r, id)
	if err != nil {
		return nil, err
	}

	r.mux.Lock()
	defer r.mux.Unlock()
	r.peers[id] = peer

	for _, forwarder := range r.forwarders {
		local, err := forwarder.AddPeer(id)
		if err != nil {
			continue
		}

		peer.conn.AddTrack(local)
	}

	// Don't need to renegotiate here; negotiation is not over yet
	//if err := peer.Renegotiate(); err != nil {
	//	return nil, err
	//}

	return peer, nil
}

func (r *Room) addIncomingTrack(from *Peer, remote *webrtc.TrackRemote) {
	r.mux.Lock()
	defer r.mux.Unlock()

	forwarder := NewTrackForwarder(remote)
	r.forwarders[remote.ID()] = forwarder

	for peerID, peer := range r.peers {
		if peerID == from.ID {
			continue
		}

		local, err := forwarder.AddPeer(peerID)
		if err != nil {
			continue
		}

		peer.conn.AddTrack(local)
		if err := peer.Renegotiate(); err != nil {
			continue
		}
	}

	forwarder.Start()
}
