package grpcv1

import (
	"google.golang.org/grpc"
)

type Handler struct {
}

func NewHandler(server grpc.ServiceRegistrar, s Signaling) *Handler {
	registerSignalingServer(server, s)

	return &Handler{}
}
