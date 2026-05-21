package main

import (
	"cbr/internal/app"
	"log/slog"
	"os"
)

func main() {
	a := app.New()
	if err := a.Run(); err != nil {
		slog.Error("app error", slog.Any("error", err))
		os.Exit(1)
	}
}
