package main

import (
	"log/slog"
	"os"

	"github.com/CMAK12/gonference/internal/streaming/server"
)

func main() {
	if err := server.Run(); err != nil {
		slog.Error("streaming exited with error", slog.String("error", err.Error()))
		os.Exit(1)
	}
}
