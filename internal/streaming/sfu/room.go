package sfu

import (
	"errors"
	"log/slog"
	"sync"

	"github.com/CMAK12/gonference/internal/streaming/entity"

	"github.com/pion/webrtc/v3"
)

type Room struct {
	id string

	mux        sync.RWMutex
	peers      map[string]*Peer
	forwarders map[string]*TrackForwarder

	api    *webrtc.API
	logger *slog.Logger
}

func NewRoom(api *webrtc.API, id string) *Room {
	return &Room{
		id:         id,
		api:        api,
		logger:     slog.Default().With("room_id", id),
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

func (r *Room) AddPeer(offer webrtc.SessionDescription, id string) (webrtc.SessionDescription, *Peer, error) {
	peer, err := NewPeer(r.api, r, id)
	if err != nil {
		return webrtc.SessionDescription{}, nil, err
	}

	answer, err := peer.CreateAnswer(offer)
	if err != nil {
		return webrtc.SessionDescription{}, nil, errors.Join(err, peer.Close())
	}

	r.attachPeer(peer)

	return answer, peer, nil
}

func (r *Room) attachPeer(peer *Peer) {
	id := peer.ID()

	r.mux.Lock()
	replaced := r.peers[id]
	r.peers[id] = peer
	forwarders := make([]*TrackForwarder, 0, len(r.forwarders))
	for _, f := range r.forwarders {
		forwarders = append(forwarders, f)
	}
	r.mux.Unlock()

	if replaced != nil {
		if err := replaced.Close(); err != nil {
			replaced.logger.Error("Failed to close replaced peer", slog.String("error", err.Error()))
		}
	}

	for _, forwarder := range forwarders {
		local, err := forwarder.AddPeer(id)
		if err != nil {
			peer.logger.Error("Failed to add peer to forwarder", slog.String("error", err.Error()))
			continue
		}

		if _, err := peer.addOutboundTrack(local); err != nil {
			peer.logger.Error("Failed to add track to peer connection", slog.String("error", err.Error()))
		}
	}
}

func (r *Room) RemovePeer(id string) {
	peer, ok := r.GetPeer(id)
	if !ok {
		return
	}

	r.removePeer(peer)
}

func (r *Room) removePeer(peer *Peer) {
	id := peer.ID()

	r.mux.Lock()
	if r.peers[id] != peer {
		r.mux.Unlock()
		return
	}
	delete(r.peers, id)

	var dead []*TrackForwarder
	for trackID, forwarder := range r.forwarders {
		if forwarder.peer == peer {
			dead = append(dead, forwarder)
			delete(r.forwarders, trackID)
			continue
		}
		forwarder.RemovePeer(id)
	}

	if len(r.peers) == 0 {
		r.mux.Unlock()
		r.Close()
	}

	otherPeers := make([]*Peer, 0, len(r.peers))
	for _, p := range r.peers {
		otherPeers = append(otherPeers, p)
	}
	r.mux.Unlock()

	for _, forwarder := range dead {
		forwarder.Close()
	}

	if err := peer.Close(); err != nil {
		peer.logger.Error("Failed to close peer", slog.String("error", err.Error()))
	}

	for _, p := range otherPeers {
		if err := p.removeOutboundTracks(dead); err != nil {
			p.logger.Error("Failed to remove outbound tracks", slog.String("error", err.Error()))
		}
	}

	leaveMsg := entity.SignalMessage{
		Type:     entity.TypeLeave,
		RoomID:   r.id,
		MemberID: id,
	}

	r.broadcast(leaveMsg, peer)
}

func (r *Room) broadcast(msg entity.SignalMessage, except *Peer) {
	r.mux.RLock()
	peers := make([]*Peer, 0, len(r.peers))
	for _, p := range r.peers {
		if p != except {
			peers = append(peers, p)
		}
	}
	r.mux.RUnlock()

	for _, peer := range peers {
		if err := peer.sendSignal(msg); err != nil {
			peer.logger.Error("Failed to send signal message", slog.String("error", err.Error()))
		}
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
