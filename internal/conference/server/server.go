package server

import (
	"context"
	"fmt"
	"log/slog"
	"os/signal"
	"syscall"

	"github.com/CMAK12/gonference/infra/db/in-memory/dragonfly"
	"github.com/CMAK12/gonference/infra/db/relational/postgres"
	infraserver "github.com/CMAK12/gonference/infra/server/grpc"
	"github.com/CMAK12/gonference/internal/conference/config"
	grpcv1 "github.com/CMAK12/gonference/internal/conference/handler/grpc/v1"
	"github.com/CMAK12/gonference/internal/conference/service"
	in_memory "github.com/CMAK12/gonference/internal/conference/storage/in-memory"
	"github.com/CMAK12/gonference/internal/conference/storage/relational"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/keepalive"
)

const (
	serviceName       = "conference"
	healthServiceName = "CONFERENCE_V1"
)

func Run() error {
	log := slog.Default().With(slog.String("service", serviceName))

	cfg := config.MustLoad()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	df, err := dragonfly.NewClient(ctx, cfg.Dragonfly)
	if err != nil {
		return fmt.Errorf("connect dragonfly: %w", err)
	}

	pg, err := postgres.New(ctx, cfg.Postgres)
	if err != nil {
		return fmt.Errorf("connect postgres: %w", err)
	}
	defer pg.Close()

	svc := service.NewService(in_memory.NewStorage(df), relational.NewUnitOfWork(pg))

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
			return fmt.Errorf("serve conference: %w", err)
		}

		log.Info("conference stopped")

		return nil
	case <-ctx.Done():
		log.Info("shutdown signal received, draining conference")
	}

	grpcServer.SetServingStatus(healthServiceName, grpc_health_v1.HealthCheckResponse_NOT_SERVING)

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.GRPC.ShutdownTimeout)
	defer cancel()

	if err := grpcServer.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("shutdown conference: %w", err)
	}

	if err := <-serveErr; err != nil {
		return fmt.Errorf("serve conference: %w", err)
	}

	log.Info("conference stopped")

	return nil
}
