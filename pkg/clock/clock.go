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

func (
	Clock,
) WeekStart(
	value time.Time,
) time.Time {
	day := int(value.Weekday())
	if day == 0 {
		day = 7
	}
	start := value.AddDate(0, 0, -(day - 1))
	return DayStart(start)
}

func (
	c Clock,
) WeekEnd(
	value time.Time,
) time.Time {
	return c.WeekStart(value).AddDate(0, 0, 7).Add(-time.Nanosecond)
}
