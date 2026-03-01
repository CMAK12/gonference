package server

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"gonference/internal/config"
	"gonference/internal/controller/admin_panel"
	"gonference/internal/controller/rest"
	"gonference/internal/usecase"
)

func Run() {
	cfg := config.MustLoad()

	uc, err := usecase.NewUseCase()
	if err != nil {
		slog.Error("Failed to create UseCase", slog.String("error", err.Error()))
	}

	rest := rest.NewHandler(cfg.REST, uc)
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
