package core

//import (
//	"fmt"
//	"io"
//	"sync"
//
//	"gonference/pkg/conference"
//
//	"github.com/pion/webrtc/v3"
//)
//
//type Hub struct {
//	mux   sync.RWMutex
//	rooms map[string]map[string]*conference.Member // roomID -> memberID -> Member
//}
//
//func NewHub() *Hub {
//	return &Hub{
//		rooms: make(map[string]map[string]*conference.Member),
//	}
//}
//
//func (h *Hub) AddPeer(roomID, memberID string, member *conference.Member) {
//	h.mux.Lock()
//	defer h.mux.Unlock()
//
//	if _, ok := h.rooms[roomID]; !ok {
//		h.rooms[roomID] = make(map[string]*conference.Member)
//	}
//
//	member.OnClose(func() {
//		h.RemovePeer(roomID, memberID)
//	})
//
//	member.SetTrackHandler(func(remote *webrtc.TrackRemote, receiver *webrtc.RTPReceiver) {
//		fmt.Printf("Starting track handler for member %s, track kind: %s, streamID: %s\n",
//			memberID, remote.Kind().String(), remote.StreamID())
//
//		go func() {
//			for {
//				pkt, _, err := remote.ReadRTP()
//				if err != nil {
//					if err == io.EOF {
//						fmt.Printf("Track ended for member %s\n", memberID)
//						return
//					}
//
//					fmt.Printf("Error reading RTP packet: %v\n", err)
//					return
//				}
//
//				h.mux.RLock()
//				peers, ok := h.rooms[roomID]
//				h.mux.RUnlock()
//
//				if !ok {
//					return
//				}
//
//				for id, peer := range peers {
//					// Skip sending to self
//					if id == memberID {
//						continue
//					}
//
//					if err := peer.WriteRTP(pkt); err != nil {
//						fmt.Printf("Error writing RTP packet from %s to peer %s: %v\n", memberID, id, err)
//					}
//				}
//			}
//		}()
//	})
//
//	h.rooms[roomID][memberID] = member
//}
//
//func (h *Hub) RemovePeer(roomID, memberID string) {
//	h.mux.Lock()
//	defer h.mux.Unlock()
//
//	if peers, ok := h.rooms[roomID]; ok {
//		delete(peers, memberID)
//		if len(peers) == 0 {
//			delete(h.rooms, roomID)
//		}
//	}
//}
