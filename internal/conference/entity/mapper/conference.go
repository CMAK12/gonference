package mapper

import (
	"strings"

	"github.com/CMAK12/gonference/internal/conference/entity"
	pb "github.com/CMAK12/gonference/internal/gen/conference/v1"
)

// invitedMembersSep joins the repeated proto invited_members field into the
// single string the entity/storage layer uses.
const invitedMembersSep = ","

// CreateRequestToConference maps an incoming CreateConference gRPC request onto a
// domain entity. Identifier, token and creation time are assigned by the service.
func CreateRequestToConference(req *pb.CreateConferenceRequest) *entity.Conference {
	conf := &entity.Conference{
		Name:           req.GetName(),
		CreatorID:      req.GetCreatorId(),
		InvitedMembers: strings.Join(req.GetInvitedMembers(), invitedMembersSep),
	}

	if ts := req.GetStartTime(); ts != nil {
		conf.StartTime = ts.AsTime()
	}
	if ts := req.GetEndTime(); ts != nil {
		conf.EndTime = ts.AsTime()
	}
	if ts := req.GetCreatedAt(); ts != nil {
		conf.CreatedAt = ts.AsTime()
	}

	return conf
}
