package app

import (
	"bonds-report-service/internal/adapters/outbound/repository/postgreSQL"
	"bonds-report-service/internal/application/ports"
	"bonds-report-service/internal/closer"
	"context"
)

func (d *diContainer) Storage() ports.Storage {
	if d.storage == nil {
		ctx, cancel := context.WithTimeout(context.Background(), d.config.Timeouts.DBConnectTimeout)
		defer cancel()

		d.logger.Info("create postgres pool")
		pool, err := postgreSQL.NewPool(ctx, d.config)
		if err != nil {
			d.logger.Error("failed to create PostgreSQL pool", "err", err)
			panic(err)
		}

		d.logger.Info("create postgres storage")
		serviceStorage := postgreSQL.NewStorage(d.logger, pool)

		d.logger.Info("PostgreSQL storage initialized successfully")

		d.storage = serviceStorage
		closer.Add("postgres DB", func(context.Context) error {
			return serviceStorage.Close()
		})
	}

	return d.storage
}
