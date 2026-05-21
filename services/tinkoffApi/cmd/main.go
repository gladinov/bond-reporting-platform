package main

import (
	"log/slog"
	"os"
	"tinkoffApi/internal/app"
)

func main() {
	a := app.New()
	if err := a.Run(); err != nil {
		slog.Error("app error", slog.Any("error", err))
		os.Exit(1)
	}
}
