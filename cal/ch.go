package ch

import (
	"github.com/rickar/cal/v2"
	"time"
)

var (
	Bettag = &cal.Holiday{
		Name:    "Eidgenössischer Dank-, Buss- und Bettag",
		Weekday: time.Sunday,
		Month:   time.September,
		Offset:  3,
		Type:    cal.ObservanceReligious,
		Func:    cal.CalcWeekdayFrom,
	}

	BettagMontag = &cal.Holiday{
		Name:    "Eidgenössischer Dank-, Buss- und Bettag Montag",
		Weekday: time.Sunday,
		Month:   time.September,
		Offset:  3,
		Type:    cal.ObservancePublic,
		Func: func(h *cal.Holiday, year int) time.Time {
			t := cal.CalcWeekdayFrom(h, year)
			return t.AddDate(0, 0, 1)
		},
	}
)
