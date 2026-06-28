package server

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	streamingclient "github.com/CMAK12/gonference/infra/client/streaming"
	"github.com/CMAK12/gonference/internal/gateway/config"
	"github.com/CMAK12/gonference/internal/gateway/rest"
)

func Run() {
	cfg := config.MustLoad()

	streaming, err := streamingclient.New(cfg.Streaming.Addr())
	if err != nil {
		slog.Error("Failed to create streaming client", slog.String("error", err.Error()))
		os.Exit(1)
	}

	api := rest.NewHandler(cfg.REST, streaming)
	go api.ListenAndServe()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)

	sig := <-sigChan
	slog.Info("Execution interrupted", slog.String("signal", sig.String()))

	api.Close()
	if err := streaming.Close(); err != nil {
		slog.Error("Failed to close streaming client", slog.String("error", err.Error()))
	}
}
