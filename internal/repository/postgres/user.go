package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepo struct {
	db *pgxpool.Pool
}

func NewUserRepo(db *pgxpool.Pool) *UserRepo {
	return &UserRepo{db: db}
}

func (r *UserRepo) GetSalary(ctx context.Context, telegramID int64) (float64, bool, error) {
	var salary float64
	err := r.db.QueryRow(ctx,
		`SELECT salary FROM users WHERE telegram_id = $1`,
		telegramID,
	).Scan(&salary)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, false, nil
	}
	if err != nil {
		return 0, false, err
	}
	return salary, true, nil
}

func (r *UserRepo) SetSalary(ctx context.Context, telegramID int64, salary float64) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO users (telegram_id, salary, updated_at)
		 VALUES ($1, $2, NOW())
		 ON CONFLICT (telegram_id) DO UPDATE SET salary = EXCLUDED.salary, updated_at = NOW()`,
		telegramID, salary,
	)
	return err
}
