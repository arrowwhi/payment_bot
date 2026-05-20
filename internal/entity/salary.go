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
	Total            float64
	TotalWorkingDays int
}
