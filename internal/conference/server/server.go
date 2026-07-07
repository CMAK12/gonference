package server

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	infraserver "github.com/CMAK12/gonference/infra/server/grpc"
	"github.com/CMAK12/gonference/internal/conference/config"
	grpcv1 "github.com/CMAK12/gonference/internal/streaming/handler/grpc/v1"
)

func Run() {
	// DEPENDENCIES

	slog.Info("Starting conference service")

	cfg := config.MustLoad()

	grpcServer, err := infraserver.New(cfg.GRPC.Address)
	if err != nil {
		slog.Error("Failed to create gRPC server", "error", err)
		return
	}

	// TODO: Replace 'nil' with your actual ConferenceService implementation
	_ = grpcv1.NewHandler(grpcServer, nil)

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

	sig := <-sigChan
	slog.Info("Execution interrupted, shutting down streaming service", slog.String("signal", sig.String()))

	grpcServer.GracefulStop()

	slog.Info("Streaming service stopped")
}
