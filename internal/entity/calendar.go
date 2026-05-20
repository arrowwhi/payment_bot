package entity

import "time"

type WorkdayCalendar struct {
	Year  int
	Month time.Month
	Days  []bool // index 0 = day 1; true = working day
}

func (c *WorkdayCalendar) CountWorkingDays(from, to int) int {
	count := 0
	for i := from - 1; i < to && i < len(c.Days); i++ {
		if c.Days[i] {
			count++
		}
	}
	return count
}

func (c *WorkdayCalendar) TotalWorkingDays() int {
	return c.CountWorkingDays(1, len(c.Days))
}
