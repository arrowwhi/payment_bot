package usecase

import "math"

type taxBracket struct {
	upTo float64 // верхняя граница годового дохода, включительно
	rate float64
}

// progressiveScale возвращает шкалу НДФЛ (ст. 224 НК РФ): пятиступенчатая с 2025 года, двухступенчатая до него.
func progressiveScale(year int) []taxBracket {
	if year >= 2025 {
		return []taxBracket{
			{2_400_000, 0.13},
			{5_000_000, 0.15},
			{20_000_000, 0.18},
			{50_000_000, 0.20},
			{math.Inf(1), 0.22},
		}
	}
	return []taxBracket{
		{5_000_000, 0.13},
		{math.Inf(1), 0.15},
	}
}

// annualTax возвращает налог с дохода, накопленного с начала года.
func annualTax(income float64, year int) float64 {
	var tax, lower float64
	for _, b := range progressiveScale(year) {
		if income <= lower {
			break
		}
		tax += (math.Min(income, b.upTo) - lower) * b.rate
		lower = b.upTo
	}
	return tax
}

// incomeTax возвращает НДФЛ с дохода, полученного после того, как с начала года уже накоплено prevIncome.
// Принимает накопленный доход, а не месяц, чтобы отпускные и прочие выплаты можно было учитывать отдельно от зарплаты.
func incomeTax(prevIncome, income float64, year int) float64 {
	return round2(annualTax(prevIncome+income, year) - annualTax(prevIncome, year))
}
