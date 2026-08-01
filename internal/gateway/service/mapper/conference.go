package mapper

import (
	confpb "github.com/CMAK12/gonference/internal/gen/conference/v1"
	gatewaypb "github.com/CMAK12/gonference/internal/gen/gateway/v1"
)

func ToConferenceCreateRequest(req *gatewaypb.CreateConferenceRequest) *confpb.CreateConferenceRequest {
	return &confpb.CreateConferenceRequest{
		InvitedMembers: req.GetInvitedMembers(),
		StartTime:      req.GetStartTime(),
		EndTime:        req.GetEndTime(),
	}
}

func FromConferenceCreateResponse(resp *confpb.CreateConferenceResponse) *gatewaypb.CreateConferenceResponse {
	return &gatewaypb.CreateConferenceResponse{
		ConferenceId: resp.GetConferenceId(),
		Url:          resp.GetUrl(),
	}
}

func ToConferenceJoinRequest(req *gatewaypb.JoinConferenceRequest) *confpb.JoinConferenceRequest {
	return &confpb.JoinConferenceRequest{
		ConferenceId: req.GetConferenceId(),
		MemberId:     req.GetMemberId(),
	}
}

func FromConferenceJoinResponse(resp *confpb.JoinConferenceResponse) *gatewaypb.JoinConferenceResponse {
	return &gatewaypb.JoinConferenceResponse{
		ConferenceId: resp.GetConferenceId(),
		MemberId:     resp.GetMemberId(),
		Url:          resp.GetUrl(),
	}
}
