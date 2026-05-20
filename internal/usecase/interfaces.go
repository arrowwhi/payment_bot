package usecase

import (
	"context"
	"time"

	"payment_bot/internal/entity"
)

type CalendarRepository interface {
	GetWorkdayCalendar(ctx context.Context, year int, month time.Month) (*entity.WorkdayCalendar, error)
}

type UserRepository interface {
	GetSalary(ctx context.Context, telegramID int64) (salary float64, found bool, err error)
	SetSalary(ctx context.Context, telegramID int64, salary float64) error
}
