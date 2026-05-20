package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"payment_bot/internal/entity"
)

type CalendarCache struct {
	db *pgxpool.Pool
}

func NewCalendarCache(db *pgxpool.Pool) *CalendarCache {
	return &CalendarCache{db: db}
}

func (c *CalendarCache) Get(ctx context.Context, year int, month time.Month) (*entity.WorkdayCalendar, bool, error) {
	var data string
	err := c.db.QueryRow(ctx,
		`SELECT days_data FROM calendar_cache WHERE year = $1 AND month = $2`,
		year, int(month),
	).Scan(&data)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	return parseCalendar(year, month, data), true, nil
}

func (c *CalendarCache) Save(ctx context.Context, cal *entity.WorkdayCalendar) error {
	_, err := c.db.Exec(ctx,
		`INSERT INTO calendar_cache (year, month, days_data)
		 VALUES ($1, $2, $3)
		 ON CONFLICT (year, month) DO UPDATE SET days_data = EXCLUDED.days_data`,
		cal.Year, int(cal.Month), encodeCalendar(cal),
	)
	return err
}

func parseCalendar(year int, month time.Month, data string) *entity.WorkdayCalendar {
	days := make([]bool, len(data))
	for i, ch := range data {
		days[i] = ch == '0'
	}
	return &entity.WorkdayCalendar{Year: year, Month: month, Days: days}
}

func encodeCalendar(cal *entity.WorkdayCalendar) string {
	b := make([]byte, len(cal.Days))
	for i, working := range cal.Days {
		if working {
			b[i] = '0'
		} else {
			b[i] = '1'
		}
	}
	return string(b)
}
