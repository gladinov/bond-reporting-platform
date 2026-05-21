package main

import (
	"log/slog"
	"moex/internal/app"
	"os"
)

func main() {
	a := app.New()
	if err := a.Run(); err != nil {
		slog.Error("app error", slog.Any("error", err))
		os.Exit(1)
	}
}
