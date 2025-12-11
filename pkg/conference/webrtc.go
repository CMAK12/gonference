package conference

import (
	"log/slog"
	"sync"

	"github.com/pion/interceptor"
	"github.com/pion/rtp"
	"github.com/pion/webrtc/v3"
)

type Member struct {
	Username string
	logger   *slog.Logger

	PC *webrtc.PeerConnection

	videoTrack *webrtc.TrackLocalStaticRTP
	audioTrack *webrtc.TrackLocalStaticRTP

	handleTrack func(remote *webrtc.TrackRemote, receiver *webrtc.RTPReceiver)
}

func NewMember(username string) (*Member, error) {
	member := &Member{
		Username: username,
		logger:   slog.Default().With(slog.String("component", "peer")),
	}

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

	pc, err := api.NewPeerConnection(webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{"stun:stun.l.google.com:19302"},
			},
		},
	})
	if err != nil {
		return nil, err
	}

	videoTrack, err := webrtc.NewTrackLocalStaticRTP(
		webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeVP8},
		"video",
		"pion",
	)
	if err != nil {
		return nil, err
	}

	audioTrack, err := webrtc.NewTrackLocalStaticRTP(
		webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeOpus},
		"audio",
		"pion",
	)
	if err != nil {
		return nil, err
	}

	_, err = pc.AddTrack(videoTrack)
	if err != nil {
		return nil, err
	}

	_, err = pc.AddTrack(audioTrack)
	if err != nil {
		return nil, err
	}

	member.videoTrack = videoTrack
	member.audioTrack = audioTrack

	member.PC = pc

	var cleanupOnce sync.Once

	cleanup := func() {
		member.logger.Info("disconnected")
		_ = member.PC.Close()
		member.PC = nil
	}

	pc.OnConnectionStateChange(func(state webrtc.PeerConnectionState) {
		member.logger.Info("connection state changed", "state", state)

		switch state {
		case webrtc.PeerConnectionStateConnected:
			member.logger.Info("connected")
		case webrtc.PeerConnectionStateDisconnected,
			webrtc.PeerConnectionStateFailed,
			webrtc.PeerConnectionStateClosed:
			cleanupOnce.Do(cleanup)
			return
		}
	})

	pc.OnICEConnectionStateChange(func(state webrtc.ICEConnectionState) {
		member.logger.Info("ice connection state changed", "state", state)
	})

	pc.OnTrack(func(remote *webrtc.TrackRemote, receiver *webrtc.RTPReceiver) {
		h := member.handleTrack
		h(remote, receiver)
	})

	return member, nil
}

func (m *Member) SetTrackHandler(handler func(remote *webrtc.TrackRemote, receiver *webrtc.RTPReceiver)) {
	m.handleTrack = handler
}

func (m *Member) WriteRTP(pkt *rtp.Packet) error {
	switch pkt.PayloadType {
	case VP8:
		return m.videoTrack.WriteRTP(pkt)
	case OPUS:
		return m.audioTrack.WriteRTP(pkt)
	default:
		return nil
	}
}

func (m *Member) CreateAnswer(offer webrtc.SessionDescription) (*webrtc.SessionDescription, error) {
	gatherComplete := webrtc.GatheringCompletePromise(m.PC)
	if err := m.PC.SetRemoteDescription(offer); err != nil {
		return nil, err
	}

	answer, err := m.PC.CreateAnswer(nil)
	if err != nil {
		return nil, err
	}

	if err := m.PC.SetLocalDescription(answer); err != nil {
		return nil, err
	}

	<-gatherComplete

	return m.PC.LocalDescription(), nil
}
