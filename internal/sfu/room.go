package sfu

import (
	"log/slog"
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

func (r *Room) AddPeer(signal Signaling, offer webrtc.SessionDescription, id string) (*Peer, error) {
	peer, err := NewPeer(r.api, signal, r, offer, id)
	if err != nil {
		return nil, err
	}

	r.mux.Lock()
	r.peers[id] = peer
	forwarders := make([]*TrackForwarder, 0, len(r.forwarders))
	for _, f := range r.forwarders {
		forwarders = append(forwarders, f)
	}
	r.mux.Unlock()

	for _, forwarder := range forwarders {
		local, err := forwarder.AddPeer(id)
		if err != nil {
			peer.logger.Error("Failed to add peer to forwarder", slog.String("error", err.Error()))
			continue
		}

		peer.addOutboundTrack(local)
		if _, err := peer.conn.AddTrack(local); err != nil {
			peer.logger.Error("Failed to add track to peer connection", slog.String("error", err.Error()))
		}
	}

	if err := peer.Renegotiate(); err != nil {
		peer.logger.Error("Failed to renegotiate", slog.String("error", err.Error()))
		return nil, err
	}

	return peer, nil
}

func (r *Room) RemovePeer(id string) {
	r.mux.Lock()
	peer, ok := r.peers[id]
	if !ok {
		r.mux.Unlock()
		return
	}
	delete(r.peers, id)
	r.mux.Unlock()

	if err := peer.Close(); err != nil {
		peer.logger.Error("Failed to close peer", slog.String("error", err.Error()))
	}
}

func (r *Room) addIncomingTrack(from *Peer, remote *webrtc.TrackRemote) {
	r.mux.Lock()

	forwarder := NewTrackForwarder(from, remote)
	r.forwarders[remote.ID()] = forwarder

	peers := make(map[string]*Peer, len(r.peers))
	for peerID, peer := range r.peers {
		if peerID != from.ID() {
			peers[peerID] = peer
		}
	}
	r.mux.Unlock()

	forwarder.Start()

	for peerID, peer := range peers {
		local, err := forwarder.AddPeer(peerID)
		if err != nil {
			from.logger.Error("Failed to add peer to forwarder",
				slog.String("peerId", peerID),
				slog.String("error", err.Error()))
			continue
		}

		if err := peer.AddTrackAndRenegotiate(local); err != nil {
			from.logger.Error("Failed to renegotiate",
				slog.String("peerId", peerID),
				slog.String("error", err.Error()))
			continue
		}
	}
}

func (r *Room) Close() {
	r.mux.Lock()
	defer r.mux.Unlock()

	for _, peer := range r.peers {
		if err := peer.Close(); err != nil {
			peer.logger.Error("Failed to close peer", slog.String("error", err.Error()))
		}
	}

	for _, forwarder := range r.forwarders {
		forwarder.Close()
	}
}
