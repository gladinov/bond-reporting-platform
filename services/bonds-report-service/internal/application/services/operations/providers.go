package updateoperations

import (
	"bonds-report-service/internal/domain"
	"context"
)

//go:generate go run github.com/vektra/mockery/v2@v2.53.5 --name=TinkoffOperationsProvider
type TinkoffOperationsProvider interface {
	TinkoffGetOperations(ctx context.Context, opRequest domain.OperationsRequest) ([]domain.Operation, error)
}
