package analysis

import (
	"time"

	"dream117/internal/domain"
	"dream117/pkg/clock"
)

type CalendarDay struct {
	Date      string
	Count     int
	Intensity float64
	Emotions  []string
	Themes    []string
	IsToday   bool
	Weekday   int
}

type CalendarReport struct {
	From          string
	To            string
	Days          []CalendarDay
	RecordedDays  int
	LongestGap    int
	CurrentStreak int
}

func BuildCalendar(
	items []domain.Dream,
	from,
	to time.Time,
) CalendarReport {
	groups := map[string][]domain.Dream{}
	for i := range items {
		d := items[i]
		groups[d.DreamDate.Format("2006-01-02")] = append(groups[d.DreamDate.Format("2006-01-02")], d)
	}
	report := CalendarReport{From: from.Format("2006-01-02"), To: to.Format("2006-01-02")}
	for date := clock.DayStart(from); !date.After(to); date = date.AddDate(0, 0, 1) {
		key := date.Format("2006-01-02")
		dreams := groups[key]
		day := CalendarDay{Date: key, Weekday: int(date.Weekday()), IsToday: sameDate(date, time.Now().UTC())}
		if len(dreams) > 0 {
			report.RecordedDays++
			day.Emotions, day.Themes = []string{}, []string{}
			seenEmotion, seenTheme := map[string]bool{}, map[string]bool{}
			for i := range dreams {
				d := dreams[i]
				day.Count++
				day.Intensity += MoodScore(d.Emotion, d.Clarity)
				if !seenEmotion[string(d.Emotion)] {
					day.Emotions =
						append(day.Emotions,
							string(d.Emotion))
					seenEmotion[string(d.Emotion)] = true
				}
				for i := range d.Themes {
					theme := d.Themes[i]
					if !seenTheme[theme.Name] {
						day.Themes =
							append(day.Themes,
								theme.Name)
						seenTheme[theme.Name] = true
					}
				}
			}
			day.Intensity /= float64(len(dreams))
		}
		report.Days =
			append(report.Days,
				day)
	}
	report.LongestGap, report.CurrentStreak = calendarGaps(report.Days)
	return report
}

func calendarGaps(
	days []CalendarDay,
) (int, int) {
	longest, gap, streak := 0, 0, 0
	for index := len(days) - 1; index >= 0; index-- {
		if days[index].Count == 0 {
			gap++
			if gap > longest {
				longest = gap
			}
			if streak == 0 {
				streak = 0
			}
		} else {
			if index == len(days)-1 || streak > 0 {
				streak++
			}
			gap = 0
		}
	}
	return longest, streak
}
func sameDate(
	left,
	right time.Time,
) bool {
	return left.Year() == right.Year() && left.YearDay() == right.YearDay()
}
