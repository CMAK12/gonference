package server

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/CMAK12/gonference/infra/db/in-memory/dragonfly"
	postgres2 "github.com/CMAK12/gonference/infra/db/relational/postgres"
	infraserver "github.com/CMAK12/gonference/infra/server/grpc"
	"github.com/CMAK12/gonference/internal/conference/config"
	grpcv1 "github.com/CMAK12/gonference/internal/conference/handler/grpc/v1"
	"github.com/CMAK12/gonference/internal/conference/service"
	in_memory "github.com/CMAK12/gonference/internal/conference/storage/in-memory"
	"github.com/CMAK12/gonference/internal/conference/storage/relational"
)

func Run() {
	// DEPENDENCIES

	slog.Info("Starting conference service")

	ctx := context.Background()

	cfg := config.MustLoad()

	df, err := dragonfly.NewClient(ctx, dragonfly.Config{
		Addr:     cfg.Dragonfly.Addr,
		Password: cfg.Dragonfly.Password,
		DB:       cfg.Dragonfly.DB,
	})
	if err != nil {
		slog.Error("Failed to connect to dragonfly server", "error", err)

		return
	}

	postgres, err := postgres2.New(ctx, postgres2.Config{
		Host:            cfg.Postgres.Host,
		Port:            cfg.Postgres.Port,
		User:            cfg.Postgres.User,
		Password:        cfg.Postgres.Password,
		Database:        cfg.Postgres.Database,
		SSLMode:         cfg.Postgres.SSLMode,
		MaxConns:        cfg.Postgres.MaxConns,
		MinConns:        cfg.Postgres.MinConns,
		MaxConnLifetime: cfg.Postgres.MaxConnLifetime,
		MaxConnIdleTime: cfg.Postgres.MaxConnIdleTime,
	})
	if err != nil {
		slog.Error("Failed to connect to postgres server", "error", err)

		return
	}

	inMemory := in_memory.NewStorage(df)

	pg := relational.NewUnitOfWork(postgres)

	svc := service.NewService(inMemory, pg)

	grpcServer, err := infraserver.New(cfg.GRPC.Address)
	if err != nil {
		slog.Error("Failed to create gRPC server", slog.String("error", err.Error()))

		return
	}

	grpcv1.RegisterGRPCV1Handler(grpcServer, svc)

	// STARTUP

	go func() {
		if err := grpcServer.Serve(); err != nil {
			slog.Error("Failed to serve conference server", slog.String("error", err.Error()))

			return
		}
	}()

	// STOPPING

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)

	recSig := <-sigChan
	slog.Info("Execution interrupted, shutting down conference service", slog.String("signal", recSig.String()))

	grpcServer.GracefulStop()

	slog.Info("Conference service stopped")
}
