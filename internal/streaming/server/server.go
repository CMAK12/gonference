package server

import (
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	infraserver "github.com/CMAK12/gonference/infra/server/grpc"
	"github.com/CMAK12/gonference/internal/streaming/config"
	grpcv1 "github.com/CMAK12/gonference/internal/streaming/handler/grpc/v1"
	"github.com/CMAK12/gonference/internal/streaming/sfu"
	"github.com/CMAK12/gonference/internal/streaming/signaling"
)

func Run() {
	// DEPENDENCIES

	slog.Info("Starting streaming service")

	cfg := config.MustLoad()

	s, err := sfu.New()
	if err != nil {
		slog.Error("Failed to create SFU", slog.String("error", err.Error()))
		return
	}

	sig := signaling.New(s)

	addr := fmt.Sprintf(":%d", cfg.GRPC.Port)
	grpcServer, err := infraserver.New(addr)
	if err != nil {
		slog.Error("Failed to create infraserver", slog.String("error", err.Error()))
		return
	}

	_ = grpcv1.NewHandler(grpcServer, sig)

	// STARTUP

	go func() {
		if err := grpcServer.Serve(); err != nil {
			slog.Error("Failed to serve server", slog.String("error", err.Error()))
			return
		}
	}()

	// CLOSING

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)

	rsig := <-sigChan
	slog.Info("Execution interrupted, shutting down streaming service", slog.String("signal", rsig.String()))

	grpcServer.GracefulStop()
	s.Close()

	slog.Info("Streaming service stopped")
}
