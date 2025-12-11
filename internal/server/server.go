package server

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"gonference/internal/config"
	"gonference/internal/controller/admin_panel"
	"gonference/internal/controller/rest"
	"gonference/internal/core"
)

func Run() {
	cfg := config.MustLoad()

	hub := core.NewHub()

	rest := rest.NewHandler(cfg.REST, hub)
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
