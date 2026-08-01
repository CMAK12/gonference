package grpcv1

import "google.golang.org/grpc"

func RegisterGRPCV1Handler(server grpc.ServiceRegistrar, s Signaling) {
	registerSignalingServer(server, s)
}
