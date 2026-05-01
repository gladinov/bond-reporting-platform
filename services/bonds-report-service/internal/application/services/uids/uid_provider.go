package uidprovider

import (
	"bonds-report-service/internal/application/ports"
	"bonds-report-service/internal/domain"
	"context"
	"errors"
	"time"

	"github.com/gladinov/e"
)

const (
	HoursToUpdate = 12 * time.Hour
)

type UidProvider struct {
	storage                ports.Storage
	analyticsTinkoffClient ports.TinkoffAnalyticsClient
	hoursToUpdate          time.Duration
	now                    func() time.Time
}

func NewUidProvider(storage ports.Storage, analyticClient ports.TinkoffAnalyticsClient) *UidProvider {
	return &UidProvider{
		storage:                storage,
		analyticsTinkoffClient: analyticClient,
		hoursToUpdate:          HoursToUpdate,
		now:                    time.Now,
	}
}

func (u *UidProvider) GetUid(ctx context.Context, instrumentUid string) (string, error) {
	date, err := u.storage.IsUpdatedUids(ctx)
	if err != nil && !errors.Is(err, domain.ErrEmptyUids) {
		return "", e.WrapIfErr("check updated uids", err)
	}

	if errors.Is(err, domain.ErrEmptyUids) {
		return u.UpdateAndGetUid(ctx, instrumentUid)
	}

	if u.now().Sub(date) > u.hoursToUpdate {
		return u.UpdateAndGetUid(ctx, instrumentUid)
	}

	uid, err := u.storage.GetUid(ctx, instrumentUid)
	if errors.Is(err, domain.ErrEmptyUids) {
		return "", domain.ErrEmptyUidAfterUpdate
	}
	return uid, err
}

// TODO: Потенциальная проблема конкурентности

// Если 100 горутин одновременно вызовут GetUid,и TTL истёк — ты 100 раз вызовешь UpdateAndGetUid.
// Это:
// 100 сетевых вызовов
// 100 сохранений в storage
// В реальном проде это может убить сервис.
// Решение:
// mutex
// singleflight
// или внешний coordination
// Пока это не критично, но держи в голове.

// TODO: Решение прям есть в учебнике , где memo
func (u *UidProvider) UpdateAndGetUid(ctx context.Context, instrumentUid string) (string, error) {
	allAssetUids, err := u.analyticsTinkoffClient.GetAllAssetUids(ctx)
	if err != nil {
		return "", e.WrapIfErr("failed to get all asset uids", err)
	}

	uid, exist := allAssetUids[instrumentUid]
	if !exist {
		return "", domain.ErrEmptyUidAfterUpdate
	}

	if err := u.storage.SaveUids(ctx, allAssetUids); err != nil {
		return "", e.WrapIfErr("failed to save uids to storage ", err)
	}

	return uid, nil
}
