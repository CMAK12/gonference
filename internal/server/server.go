package server

import (
	"gonference/internal/sfu"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"gonference/internal/config"
	"gonference/internal/controller/admin_panel"
	"gonference/internal/controller/rest"
)

func Run() {
	cfg := config.MustLoad()

	sfu, err := sfu.New()
	if err != nil {
		slog.Error("Failed to create SFU", slog.String("error", err.Error()))
	}

	rest := rest.NewHandler(cfg.REST, sfu)
	go rest.ListenAndServe()

	ap := admin_panel.NewHandler(cfg.AdminPanel)
	go ap.ListenAndServe()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)

	sig := <-sigChan

	slog.Info("Execution interrupted", slog.String("signal", sig.String()))

	rest.Close()
	ap.Close()
}
