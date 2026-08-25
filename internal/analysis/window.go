package analysis

import (
	"time"

	"dream117/pkg/clock"
)

type Window struct {
	From  time.Time
	To    time.Time
	Label string
}

func DefaultWindow(
	days int,
) Window {
	if days <= 0 {
		days = 30
	}
	if days > 366 {
		days = 366
	}
	to := clock.Clock{}.Now()
	from := to.AddDate(0, 0, -days+1)
	return Window{From: clock.DayStart(from), To: endDay(to), Label: daysLabel(days)}
}

func ParseWindow(
	fromValue,
	toValue string,
	fallback int,
) (Window, error) {
	if fromValue == "" || toValue == "" {
		return DefaultWindow(fallback), nil
	}
	from, err := time.Parse(
		"2006-01-02", fromValue)
	if err != nil {
		return Window{}, err
	}
	to, err := time.Parse(
		"2006-01-02", toValue)
	if err != nil {
		return Window{}, err
	}
	return Window{From: clock.DayStart(from), To: endDay(to), Label: fromValue + " 至 " + toValue}, nil
}
func endDay(
	value time.Time,
) time.Time {
	return clock.DayStart(value).Add(24*time.Hour - time.Nanosecond)
}
func daysLabel(
	days int,
) string {
	return map[int]string{7: "近 7 天", 30: "近 30 天", 90: "近 90 天"}[days]
}
