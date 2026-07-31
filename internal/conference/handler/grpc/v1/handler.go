package grpcv1

import "google.golang.org/grpc"

func RegisterGRPCV1Handler(server grpc.ServiceRegistrar, s ConferenceService) {
	registerConferenceServer(server, s)
}
