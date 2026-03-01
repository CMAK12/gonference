package grpc

import (
	"gonference/internal/entity"
	"gonference/pkg/pb"
)

type Transport struct {
	stream pb.Signaling_ConnectServer
}

func NewTransport(stream pb.Signaling_ConnectServer) *Transport {
	return &Transport{stream: stream}
}

func (t *Transport) ReadMessage() (entity.Message, error) {
	msg, err := t.stream.Recv()
	if err != nil {
		return entity.Message{}, err
	}

	message := entity.MapProtoToMessage(msg)

	return message, nil
}

func (t *Transport) WriteMessage(message entity.Message) error {
	body := &pb.SignalMessage{
		Type:     mapMessageTypeToProto(message.Type),
		RoomId:   message.RoomID,
		MemberId: message.MemberID,
		Sdp:      message.SDP,
		Candidate: func() *pb.ICECandidate {
			if message.Candidate == nil {
				return nil
			}
			idx := uint32(*message.Candidate.SDPMLineIndex)
			return &pb.ICECandidate{
				Candidate:        message.Candidate.Candidate,
				SdpMid:           message.Candidate.SDPMid,
				SdpMlineIndex:    &idx,
				UsernameFragment: message.Candidate.UsernameFragment,
			}
		}(),
	}

	return t.stream.Send(body)
}

func (t *Transport) Close() error {
	return nil
}

func mapMessageTypeToProto(t entity.MessageType) pb.MessageType {
	switch t {
	case entity.TypeOffer:
		return pb.MessageType_OFFER
	case entity.TypeAnswer:
		return pb.MessageType_ANSWER
	case entity.TypeCandidate:
		return pb.MessageType_CANDIDATE
	default:
		return pb.MessageType_LEAVE
	}
}
