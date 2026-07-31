package health

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
)

func NewServer(grpcServer grpc.ServiceRegistrar) *health.Server {
	server := health.NewServer()

	grpc_health_v1.RegisterHealthServer(grpcServer, server)

	server.SetServingStatus("GATEWAY_V1", grpc_health_v1.HealthCheckResponse_SERVING)

	return server
}
