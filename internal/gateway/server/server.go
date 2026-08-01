package server

import (
	"context"
	"fmt"
	"log/slog"
	"os/signal"
	"syscall"

	"github.com/CMAK12/gonference/infra/client/conference"
	infraserver "github.com/CMAK12/gonference/infra/server/grpc"
	"github.com/CMAK12/gonference/internal/gateway/config"
	grpcv1 "github.com/CMAK12/gonference/internal/gateway/handler/grpc/v1"
	"github.com/CMAK12/gonference/internal/gateway/service"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/keepalive"
)

const serviceName = "gateway"
const healthServiceName = "GATEWAY_V1"

func Run() error {
	log := slog.Default().With(slog.String("service", serviceName))

	cfg := config.MustLoad()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	conferenceClient, err := conference.NewClient(cfg.Client.Conference.Addr)
	if err != nil {
		return fmt.Errorf("create conference client: %w", err)
	}
	defer func() {
		if err := conferenceClient.Close(); err != nil {
			log.Error("close conference client", slog.String("error", err.Error()))
		}
	}()

	log.Info("conference client ready", "addr", cfg.Client.Conference.Addr)

	svc := service.New(conferenceClient)

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

	grpcv1.RegisterGRPCV1Handler(grpcServer, svc)

	serveErr := make(chan error, 1)

	go func() { serveErr <- grpcServer.Serve() }()

	select {
	case err := <-serveErr:
		if err != nil {
			return fmt.Errorf("serve gateway: %w", err)
		}

		log.Info("gateway stopped")

		return nil
	case <-ctx.Done():
		log.Info("shutdown signal received, draining gateway")
	}

	grpcServer.SetServingStatus(healthServiceName, grpc_health_v1.HealthCheckResponse_NOT_SERVING)

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.GRPC.ShutdownTimeout)
	defer cancel()

	if err := grpcServer.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("shutdown gateway: %w", err)
	}

	if err := <-serveErr; err != nil {
		return fmt.Errorf("serve gateway: %w", err)
	}

	log.Info("gateway stopped")

	return nil
}
