package server

import (
	"context"
	"fmt"
	"log/slog"
	"os/signal"
	"syscall"

	infraserver "github.com/CMAK12/gonference/infra/server/grpc"
	"github.com/CMAK12/gonference/internal/streaming/config"
	grpcv1 "github.com/CMAK12/gonference/internal/streaming/handler/grpc/v1"
	"github.com/CMAK12/gonference/internal/streaming/sfu"
	"github.com/CMAK12/gonference/internal/streaming/signaling"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/keepalive"
)

const (
	serviceName       = "streaming"
	healthServiceName = "STREAMING_V1"
)

func Run() error {
	log := slog.Default().With(slog.String("service", serviceName))

	cfg := config.MustLoad()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	s, err := sfu.New()
	if err != nil {
		return fmt.Errorf("create sfu: %w", err)
	}
	defer s.Close()

	sig := signaling.New(s)

	grpcServer, err := infraserver.New(cfg.GRPC.Addr,
		infraserver.WithLogger(log),
		infraserver.WithMessageSizes(cfg.GRPC.MaxRecvMsgSize, cfg.GRPC.MaxSendMsgSize),
		infraserver.WithConnectionTimeout(cfg.GRPC.ConnectionTimeout),
		infraserver.WithShutdownTimeout(cfg.GRPC.ShutdownTimeout),
		infraserver.WithKeepalive(keepalive.ServerParameters{
			MaxConnectionIdle: cfg.GRPC.MaxConnectionIdle,
			Time:              cfg.GRPC.KeepaliveTime,
			Timeout:           cfg.GRPC.KeepaliveTimeout,
		}),
		infraserver.WithKeepaliveEnforcement(keepalive.EnforcementPolicy{
			MinTime:             cfg.GRPC.MinClientPingInterval,
			PermitWithoutStream: true,
		}),
		infraserver.WithHealthCheck(true),
		infraserver.WithReflection(cfg.GRPC.Reflection),
	)
	if err != nil {
		return fmt.Errorf("create grpc server: %w", err)
	}

	grpcServer.SetServingStatus(healthServiceName, grpc_health_v1.HealthCheckResponse_SERVING)

	grpcv1.RegisterGRPCV1Handler(grpcServer, sig)

	serveErr := make(chan error, 1)

	go func() { serveErr <- grpcServer.Serve() }()

	select {
	case err := <-serveErr:
		if err != nil {
			return fmt.Errorf("serve streaming: %w", err)
		}

		log.Info("streaming stopped")

		return nil
	case <-ctx.Done():
		log.Info("shutdown signal received, draining streaming")
	}

	grpcServer.SetServingStatus(healthServiceName, grpc_health_v1.HealthCheckResponse_NOT_SERVING)

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.GRPC.ShutdownTimeout)
	defer cancel()

	if err := grpcServer.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("shutdown streaming: %w", err)
	}

	if err := <-serveErr; err != nil {
		return fmt.Errorf("serve streaming: %w", err)
	}

	log.Info("streaming stopped")

	return nil
}
