package grpcv1

import "google.golang.org/grpc"

type Handler struct {
}

func NewHandler(server grpc.ServiceRegistrar, s ConferenceService) *Handler {
	registerConferenceServer(server, s)

	return &Handler{}
}
