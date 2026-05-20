package telegram

import (
	"fmt"
	"math"
	"strconv"
	"time"

	"payment_bot/internal/entity"
)

var monthNames = map[time.Month]string{
	time.January:   "январь",
	time.February:  "февраль",
	time.March:     "март",
	time.April:     "апрель",
	time.May:       "май",
	time.June:      "июнь",
	time.July:      "июль",
	time.August:    "август",
	time.September: "сентябрь",
	time.October:   "октябрь",
	time.November:  "ноябрь",
	time.December:  "декабрь",
}

var monthNamesGenitive = map[time.Month]string{
	time.January:   "января",
	time.February:  "февраля",
	time.March:     "марта",
	time.April:     "апреля",
	time.May:       "мая",
	time.June:      "июня",
	time.July:      "июля",
	time.August:    "августа",
	time.September: "сентября",
	time.October:   "октября",
	time.November:  "ноября",
	time.December:  "декабря",
}

type Formatter struct{}

func NewFormatter() *Formatter { return &Formatter{} }

func (f *Formatter) FormatPrivate(b *entity.PaymentBreakdown) string {
	advancePeriod := fmt.Sprintf("1–15 %s", monthNamesGenitive[b.Month])
	remainderPeriod := fmt.Sprintf("16–%d %s", b.Remainder.EndDay, monthNamesGenitive[b.Month])

	return fmt.Sprintf(
		"Расчёт за %s %d\n\n"+
			"Аванс (%s):\n"+
			"  Рабочих дней: %d из %d\n"+
			"  Сумма: %s (%s%%)\n\n"+
			"Остаток (%s):\n"+
			"  Рабочих дней: %d из %d\n"+
			"  Сумма: %s (%s%%)\n\n"+
			"Итого: %s",
		monthNames[b.Month], b.Year,
		advancePeriod,
		b.Advance.WorkingDays, b.TotalWorkingDays,
		formatMoney(b.Advance.Amount), formatPct(b.Advance.Percentage),
		remainderPeriod,
		b.Remainder.WorkingDays, b.TotalWorkingDays,
		formatMoney(b.Remainder.Amount), formatPct(b.Remainder.Percentage),
		formatMoney(b.Total),
	)
}

func (f *Formatter) FormatGroup(b *entity.PaymentBreakdown) string {
	advancePeriod := fmt.Sprintf("1–15 %s", monthNamesGenitive[b.Month])
	remainderPeriod := fmt.Sprintf("16–%d %s", b.Remainder.EndDay, monthNamesGenitive[b.Month])

	return fmt.Sprintf(
		"Расчёт за %s %d\n\n"+
			"Аванс (%s): %d раб. дн. — %s%%\n"+
			"Остаток (%s): %d раб. дн. — %s%%",
		monthNames[b.Month], b.Year,
		advancePeriod, b.Advance.WorkingDays, formatPct(b.Advance.Percentage),
		remainderPeriod, b.Remainder.WorkingDays, formatPct(b.Remainder.Percentage),
	)
}

func formatMoney(amount float64) string {
	rounded := math.Round(amount*100) / 100
	intPart := int64(rounded)
	fracPart := int(math.Round((rounded - float64(intPart)) * 100))
	if fracPart < 0 {
		fracPart = -fracPart
	}

	s := strconv.FormatInt(intPart, 10)
	result := make([]byte, 0, len(s)+len(s)/3)
	for i := range len(s) {
		if i > 0 && (len(s)-i)%3 == 0 {
			result = append(result, ' ')
		}
		result = append(result, s[i])
	}
	return fmt.Sprintf("%s,%02d ₽", string(result), fracPart)
}

func formatPct(pct float64) string {
	return strconv.FormatFloat(pct, 'f', 1, 64)
}
