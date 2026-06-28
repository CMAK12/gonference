package server

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/CMAK12/gonference/internal/streaming/config"
	"github.com/CMAK12/gonference/internal/streaming/sfu"
	"github.com/CMAK12/gonference/internal/streaming/signaling"
)

func Run() {
	_ = config.MustLoad()

	s, err := sfu.New()
	if err != nil {
		slog.Error("Failed to create SFU", slog.String("error", err.Error()))
		return
	}
	_ = signaling.New(s)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)

	rsig := <-sigChan
	slog.Info("Execution interrupted", slog.String("signal", rsig.String()))

	s.Close()
}
