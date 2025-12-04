package server

import (
	"gonference/internal/config"
	"gonference/internal/controller/admin_panel"
	rest "gonference/internal/controller/rest"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

func Run() {
	cfg := config.MustLoad()

	rest := rest.NewHandler(cfg.REST)
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
