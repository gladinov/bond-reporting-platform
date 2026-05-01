package positionProcessor

import "context"

//go:generate go run github.com/vektra/mockery/v2@v2.53.5 --name=UidProvider
type UidProvider interface {
	GetUid(ctx context.Context, instrumentUid string) (string, error)
}
