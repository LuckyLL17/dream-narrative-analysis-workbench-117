package store

import (
	"sort"
	"time"

	"dream117/internal/domain"
)

type Aggregate struct {
	Count         int
	ClarityTotal  float64
	SleepTotal    float64
	Remembered    int
	EmotionCounts map[string]int
	ThemeCounts   map[string]int
	TagCounts     map[string]int
	FirstDate     time.Time
	LastDate      time.Time
}

func (
	a *Aggregate,
) Add(
	d domain.Dream,
) {
	a.Count++
	a.ClarityTotal += float64(d.Clarity)
	a.SleepTotal += d.SleepHours
	if d.RememberDetail {
		a.Remembered++
	}
	if a.EmotionCounts == nil {
		a.EmotionCounts = map[string]int{}
	}
	a.EmotionCounts[string(d.Emotion)]++
	if a.ThemeCounts == nil {
		a.ThemeCounts = map[string]int{}
	}
	for i := range d.Themes {
		theme := d.Themes[i]
		a.ThemeCounts[theme.Name]++
	}
	if a.TagCounts == nil {
		a.TagCounts = map[string]int{}
	}
	for i := range d.Tags {
		tag := d.Tags[i]
		a.TagCounts[tag]++
	}
	if a.FirstDate.IsZero() || d.DreamDate.Before(a.FirstDate) {
		a.FirstDate = d.DreamDate
	}
	if d.DreamDate.After(a.LastDate) {
		a.LastDate = d.DreamDate
	}
}

func (a Aggregate) AverageClarity() float64 {
	if a.Count == 0 {
		return 0
	}
	return a.ClarityTotal / float64(a.Count)
}

func (a Aggregate) AverageSleep() float64 {
	if a.Count == 0 {
		return 0
	}
	return a.SleepTotal / float64(a.Count)
}

func (a Aggregate) RememberRate() float64 {
	if a.Count == 0 {
		return 0
	}
	return float64(a.Remembered) / float64(a.Count)
}

func (
	s *Store,
) AggregateDreams(
	userID string,
	from,
	to time.Time,
) Aggregate {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := Aggregate{}
	for i := range s.data.Dreams {
		d := s.data.Dreams[i]
		if d.UserID != userID || d.DreamDate.Before(from) || d.DreamDate.After(to) {
			continue
		}
		result.Add(d)
	}
	return result
}

func SortedDates(
	items []domain.Dream,
) []time.Time {
	dates := make(
		[]time.Time,
		0,
		len(items),
	)
	seen := map[string]bool{}
	for i := range items {
		item := items[i]
		date := DateFloor(item.DreamDate)
		key := date.Format("2006-01-02")
		if !seen[key] {
			dates = append(dates, date)
			seen[key] = true
		}
	}
	sort.Slice(
		dates,
		func(left, right int) bool {
			return dates[left].Before(dates[right])
		},
	)
	return dates
}

func jobDueFirst(left, right domain.AnalysisJob) bool {
	return left.ScheduledAt.Before(right.ScheduledAt)
}

func newestDreamFirst(left, right domain.Dream) bool {
	return left.DreamDate.After(right.DreamDate)
}

func oldestDreamFirst(left, right domain.Dream) bool {
	return left.DreamDate.Before(right.DreamDate)
}

func newestReportFirst(left, right domain.WeeklyReport) bool {
	return left.WeekStart.After(right.WeekStart)
}
