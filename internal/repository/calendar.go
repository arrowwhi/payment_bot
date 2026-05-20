package repository

import (
	"context"
	"time"

	"payment_bot/internal/entity"
)

type calendarCacheStorage interface {
	Get(ctx context.Context, year int, month time.Month) (*entity.WorkdayCalendar, bool, error)
	Save(ctx context.Context, cal *entity.WorkdayCalendar) error
}

type calendarFetcher interface {
	FetchWorkdayCalendar(ctx context.Context, year int, month time.Month) (*entity.WorkdayCalendar, error)
}

type CachingCalendarRepository struct {
	cache   calendarCacheStorage
	fetcher calendarFetcher
}

func NewCachingCalendarRepository(cache calendarCacheStorage, fetcher calendarFetcher) *CachingCalendarRepository {
	return &CachingCalendarRepository{cache: cache, fetcher: fetcher}
}

func (r *CachingCalendarRepository) GetWorkdayCalendar(ctx context.Context, year int, month time.Month) (*entity.WorkdayCalendar, error) {
	cal, found, err := r.cache.Get(ctx, year, month)
	if err != nil {
		return nil, err
	}
	if found {
		return cal, nil
	}

	cal, err = r.fetcher.FetchWorkdayCalendar(ctx, year, month)
	if err != nil {
		return nil, err
	}

	_ = r.cache.Save(ctx, cal) // best-effort: don't fail on cache write error

	return cal, nil
}
