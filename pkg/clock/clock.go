package clock

import "time"

type Clock struct{}

func (Clock) Now() time.Time { return time.Now().UTC() }

func (Clock) Today() time.Time {
	now := time.Now().UTC()
	return DayStart(now)
}

func DayStart(value time.Time) time.Time {
	return time.Date(
		value.Year(),
		value.Month(),
		value.Day(),
		0,
		0,
		0,
		0,
		time.UTC,
	)
}

func (
	Clock,
) MonthStart(
	offset int,
) time.Time {
	now := time.Now().UTC()
	return time.Date(now.Year(), now.Month()+time.Month(offset), 1, 0, 0, 0, 0, time.UTC)
}

// WeekStart returns the Monday-00:00 UTC that begins the ISO week containing
// value. The week is Monday-based, so Sunday (time.Weekday == 0) belongs to the
// week that started on the previous Monday, not the next one.
func (
	Clock,
) WeekStart(
	value time.Time,
) time.Time {
	// time.Weekday numbers Sunday as 0. Treat Sunday as day 7 so the rollback
	// lands on the current week's Monday rather than advancing into next week.
	day := int(value.Weekday())
	if day == 0 {
		day = 7
	}
	start := value.AddDate(0, 0, -(day - 1))
	return DayStart(start)
}

// WeekEnd returns the last nanosecond (Sunday 23:59:59.999999999 UTC) of the
// Monday-based week containing value. It is the inclusive upper bound paired
// with WeekStart so the same week never yields different edges across callers.
func (
	c Clock,
) WeekEnd(
	value time.Time,
) time.Time {
	return c.WeekStart(value).AddDate(0, 0, 7).Add(-time.Nanosecond)
}
