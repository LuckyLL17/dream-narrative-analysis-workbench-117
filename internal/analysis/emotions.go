package analysis

import (
	"time"

	"dream117/internal/domain"
	"dream117/internal/text"
	"dream117/pkg/collections"
	"dream117/pkg/mathx"
)

type EmotionEngine struct{}

func (
	EmotionEngine,
) Trend(
	items []domain.Dream,
) []domain.EmotionPoint {
	points := collections.Map(items, func(d domain.Dream) domain.EmotionPoint {
		return domain.EmotionPoint{Date: d.DreamDate.Format("2006-01-02"), Emotion: d.Emotion, Intensity: mathx.Round(text.EmotionIntensity(d.Emotion, d.Clarity), 2), SleepHours: d.SleepHours, Clarity: d.Clarity}
	})
	collections.SortBy(
		points,
		func(left, right domain.EmotionPoint) bool {
			return left.Date < right.Date
		},
	)
	return points
}

func SleepBuckets(
	items []domain.Dream,
) map[string]map[string]float64 {
	result := map[string]map[string]float64{"短睡眠": {}, "适中": {}, "长睡眠": {}}
	for i := range items {
		d := items[i]
		bucket := "适中"
		if d.SleepHours < 6 {
			bucket = "短睡眠"
		}
		if d.SleepHours >= 9 {
			bucket = "长睡眠"
		}
		result[bucket]["clarity"] += float64(d.Clarity)
		result[bucket]["count"]++
		result[bucket]["intensity"] += text.EmotionIntensity(d.Emotion, d.Clarity)
	}
	for b := range result {
		values := result[b]
		if values["count"] > 0 {
			values["clarity_avg"] = mathx.Round(values["clarity"]/values["count"], 2)
			values["intensity_avg"] = mathx.Round(values["intensity"]/values["count"], 2)
		}
		result[b] = values
	}
	return result
}

func SleepClarityCorrelation(
	items []domain.Dream,
) float64 {
	return mathx.Round(mathx.Correlation(sleeps(items), clarityValues(items)), 3)
}

func EmotionCounts(
	items []domain.Dream,
) map[string]int {
	return collections.CountBy(items, func(d domain.Dream) string { return string(d.Emotion) })
}

func DateRange(
	items []domain.Dream,
) (time.Time, time.Time) {
	if len(items) == 0 {
		return time.Time{}, time.Time{}
	}
	min, max := items[0].DreamDate, items[0].DreamDate
	for i := range items[1:] {
		d := items[1:][i]
		if d.DreamDate.Before(min) {
			min = d.DreamDate
		}
		if d.DreamDate.After(max) {
			max = d.DreamDate
		}
	}
	return min, max
}
