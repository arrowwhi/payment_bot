package usecase

import (
	"context"
	"fmt"
	"math"
	"time"

	"payment_bot/internal/entity"
)

type SalaryCalculator struct {
	calendar CalendarRepository
}

func NewSalaryCalculator(calendar CalendarRepository) *SalaryCalculator {
	return &SalaryCalculator{calendar: calendar}
}

func (s *SalaryCalculator) Calculate(ctx context.Context, monthlySalary float64, year int, month time.Month) (*entity.PaymentBreakdown, error) {
	cal, err := s.calendar.GetWorkdayCalendar(ctx, year, month)
	if err != nil {
		return nil, fmt.Errorf("get calendar: %w", err)
	}

	daysInMonth := len(cal.Days)
	totalWorkDays := cal.TotalWorkingDays()
	if totalWorkDays == 0 {
		return nil, fmt.Errorf("no working days in %s %d", month, year)
	}

	advanceDays := cal.CountWorkingDays(1, 15)
	remainderDays := cal.CountWorkingDays(16, daysInMonth)

	// Доход с начала года считаем как зарплата × число прошедших месяцев.
	prevIncome := monthlySalary * float64(month-1)
	tax := incomeTax(prevIncome, monthlySalary, year)
	net := round2(monthlySalary - tax)

	advanceAmount := round2(net * float64(advanceDays) / float64(totalWorkDays))
	remainderAmount := round2(net - advanceAmount)

	advancePct := round2(float64(advanceDays) / float64(totalWorkDays) * 100)
	remainderPct := round2(100 - advancePct)

	return &entity.PaymentBreakdown{
		Year:  year,
		Month: month,
		Advance: entity.Period{
			Name:        "Аванс",
			StartDay:    1,
			EndDay:      15,
			WorkingDays: advanceDays,
			Amount:      advanceAmount,
			Percentage:  advancePct,
		},
		Remainder: entity.Period{
			Name:        "Остаток",
			StartDay:    16,
			EndDay:      daysInMonth,
			WorkingDays: remainderDays,
			Amount:      remainderAmount,
			Percentage:  remainderPct,
		},
		Gross:            monthlySalary,
		Tax:              tax,
		EffectiveTaxRate: round2(tax / monthlySalary * 100),
		Total:            net,
		TotalWorkingDays: totalWorkDays,
	}, nil
}

func round2(v float64) float64 {
	return math.Round(v*100) / 100
}
