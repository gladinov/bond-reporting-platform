package main

import (
	config "bonds-report-service/internal/configs"
	"bonds-report-service/internal/migrator"
	"log/slog"
	"os"

	sl "github.com/gladinov/mylogger"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	cfg := config.MustInitMigratorConfig()

	logger := sl.NewLogger(cfg.Env)

	err := migrator.Migrate(logger, cfg)
	if err != nil {
		logger.Error("migrate", slog.Any("error", err))
		os.Exit(1)
	}

	logger.Info("migrations postgres applied")
}
