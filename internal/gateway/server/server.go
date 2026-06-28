package server

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/CMAK12/gonference/internal/gateway/admin"
	"github.com/CMAK12/gonference/internal/gateway/config"
	"github.com/CMAK12/gonference/internal/gateway/rest"
)

func Run() {
	cfg := config.MustLoad()

	api := rest.NewHandler(cfg.REST)
	go api.ListenAndServe()

	ap := admin.NewHandler(cfg.Admin)
	go ap.ListenAndServe()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)

	sig := <-sigChan
	slog.Info("Execution interrupted", slog.String("signal", sig.String()))

	api.Close()
	ap.Close()
}
