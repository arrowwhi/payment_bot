package entity

import "time"

type Period struct {
	Name        string
	StartDay    int
	EndDay      int
	WorkingDays int
	Amount      float64
	Percentage  float64
}

type PaymentBreakdown struct {
	Year             int
	Month            time.Month
	Advance          Period
	Remainder        Period
	Gross            float64 // зарплата до вычета НДФЛ
	Tax              float64 // НДФЛ за месяц
	EffectiveTaxRate float64 // Tax / Gross, в процентах
	Total            float64 // сумма на руки за месяц
	TotalWorkingDays int
}
