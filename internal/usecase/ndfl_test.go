package usecase

import (
	"math"
	"testing"
)

func TestAnnualTax(t *testing.T) {
	tests := []struct {
		name   string
		income float64
		year   int
		want   float64
	}{
		{"zero", 0, 2026, 0},
		{"first bracket border", 2_400_000, 2026, 312_000},
		{"into second bracket", 2_500_000, 2026, 327_000},
		{"second bracket border", 5_000_000, 2026, 312_000 + 390_000},
		{"top bracket", 60_000_000, 2026, 312_000 + 390_000 + 2_700_000 + 6_000_000 + 2_200_000},
		{"old scale first bracket", 5_000_000, 2024, 650_000},
		{"old scale second bracket", 5_100_000, 2024, 665_000},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := annualTax(tt.income, tt.year); math.Abs(got-tt.want) > 0.001 {
				t.Errorf("annualTax(%v, %d) = %v, want %v", tt.income, tt.year, got, tt.want)
			}
		})
	}
}

func TestIncomeTax(t *testing.T) {
	tests := []struct {
		name         string
		prev, income float64
		year         int
		want         float64
	}{
		{"all at 13%", 0, 300_000, 2026, 39_000},
		{"crossing threshold mid-month", 2_300_000, 300_000, 2026, 43_000},
		{"november, salary 230k", 2_300_000, 230_000, 2026, 32_500},
		{"fully in second bracket", 2_700_000, 300_000, 2026, 45_000},
		{"old scale", 0, 300_000, 2024, 39_000},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := incomeTax(tt.prev, tt.income, tt.year); math.Abs(got-tt.want) > 0.001 {
				t.Errorf("incomeTax(%v, %v, %d) = %v, want %v", tt.prev, tt.income, tt.year, got, tt.want)
			}
		})
	}
}
