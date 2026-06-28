package server

import (
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"

	impb "github.com/CMAK12/gonference/internal/gen/streaming/v1"
	"github.com/CMAK12/gonference/internal/streaming/config"
	grpcv1 "github.com/CMAK12/gonference/internal/streaming/handler/grpc/v1"
	"github.com/CMAK12/gonference/internal/streaming/sfu"
	"github.com/CMAK12/gonference/internal/streaming/signaling"

	"google.golang.org/grpc"
)

func Run() {
	cfg := config.MustLoad()

	s, err := sfu.New()
	if err != nil {
		slog.Error("Failed to create SFU", slog.String("error", err.Error()))
		return
	}

	sig := signaling.New(s)

	grpcServer := grpc.NewServer()
	impb.RegisterSignalingServer(grpcServer, grpcv1.NewSignalingServer(sig))

	addr := fmt.Sprintf(":%d", cfg.GRPC.Port)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		slog.Error("Failed to listen", slog.String("addr", addr), slog.String("error", err.Error()))
		s.Close()
		return
	}

	go func() {
		slog.Info("gRPC signaling server started", slog.String("addr", addr))
		if err := grpcServer.Serve(lis); err != nil {
			slog.Error("gRPC server stopped", slog.String("error", err.Error()))
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)

	rsig := <-sigChan
	slog.Info("Execution interrupted", slog.String("signal", rsig.String()))

	grpcServer.GracefulStop()
	s.Close()
}
