package mapper

import (
	impb "github.com/CMAK12/gonference/internal/gen/streaming/v1"
	"github.com/CMAK12/gonference/internal/streaming/entity"
)

func MapProtoToRelayOffer(req *impb.ConnectRelayRequest) entity.RelayOffer {
	return entity.RelayOffer{
		RoomID:  req.RoomId,
		RelayID: req.RelayId,
		SDP:     req.Sdp,
	}
}

func MapRelayAnswerToProto(answer entity.RelayAnswer) *impb.ConnectRelayResponse {
	return &impb.ConnectRelayResponse{
		RoomId:  answer.RoomID,
		RelayId: answer.RelayID,
		Sdp:     answer.SDP,
	}
}
