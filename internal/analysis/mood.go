package analysis

import (
	"fmt"

	"dream117/internal/domain"
	"dream117/internal/text"
	"dream117/pkg/collections"
	"dream117/pkg/mathx"
)

type MoodObservation struct {
	Date       string
	Emotion    string
	Score      float64
	Bucket     string
	Clarity    int
	SleepHours float64
	Themes     []string
}

type MoodReport struct {
	Dominant       string
	DominantShare  float64
	Observations   []MoodObservation
	Buckets        map[string]int
	Transitions    map[string]int
	EmotionCounts  map[string]int
	AverageScore   float64
	PeakDate       string
	PeakEmotion    string
	Interpretation []string
}

var moodScores = map[domain.Emotion]float64{
	domain.EmotionCalm: 0.28, domain.EmotionConfused: 0.52, domain.EmotionHappy: 0.62,
	domain.EmotionSurprised: 0.76, domain.EmotionTense: 0.84, domain.EmotionAfraid: 0.96,
}

func BuildMoodReport(
	items []domain.Dream,
) MoodReport {
	report := MoodReport{Buckets: collections.Counter(), Transitions: collections.Counter(), EmotionCounts: collections.Counter()}
	if len(items) == 0 {
		return report
	}
	ordered := append([]domain.Dream(nil), items...)
	collections.SortBy(
		ordered,
		func(left, right domain.Dream) bool {
			return left.DreamDate.Before(right.DreamDate)
		},
	)
	var total float64
	for index := range ordered {
		d := ordered[index]
		score := MoodScore(d.Emotion, d.Clarity)
		bucket := text.MoodBucket(d.Emotion)
		report.Observations =
			append(report.Observations,
				MoodObservation{Date: d.DreamDate.Format("2006-01-02"), Emotion: string(d.Emotion), Score: score, Bucket: bucket, Clarity: d.Clarity, SleepHours: d.SleepHours, Themes: d.ThemeNames()})
		report.EmotionCounts[string(d.Emotion)]++
		report.Buckets[bucket]++
		total += score
		if index > 0 {
			key := string(ordered[index-1].Emotion) + " -> " + string(d.Emotion)
			report.Transitions[key]++
		}
	}
	report.AverageScore =
		mathx.Round(
			total/float64(len(ordered)), 3)
	report.Dominant, report.DominantShare = dominantWithShare(report.EmotionCounts, len(ordered))
	report.PeakDate, report.PeakEmotion = peakMood(report.Observations)
	report.Interpretation =
		moodInterpretation(
			report)
	return report
}

func MoodScore(
	emotion domain.Emotion,
	clarity int,
) float64 {
	base := moodScores[emotion]
	if base == 0 {
		base = 0.5
	}
	if clarity < 1 {
		clarity = 1
	}
	if clarity > 10 {
		clarity = 10
	}
	return mathx.Round(base*(0.55+float64(clarity)/20), 3)
}

func dominantWithShare(
	counts map[string]int,
	total int,
) (string, float64) {
	name, count := "", 0
	for k, n := range counts {
		if n > count {
			name, count = k, n
		}
	}
	return name, mathx.Round(mathx.Ratio(count, total), 3)
}

func peakMood(
	points []MoodObservation,
) (string, string) {
	peak := MoodObservation{}
	for i := range points {
		point := points[i]
		if point.Score > peak.Score {
			peak = point
		}
	}
	return peak.Date, peak.Emotion
}

func moodInterpretation(
	report MoodReport,
) []string {
	result := []string{}
	if report.Dominant != "" {
		result = append(result,
			"当前窗口最常见的醒来情绪是“"+report.Dominant+"”，占比约 "+formatPercent(report.DominantShare)+"。")
	}
	if report.PeakDate != "" {
		result = append(result,
			"情绪强度最高的一次出现在 "+report.PeakDate+"，标签为“"+report.PeakEmotion+"”。")
	}
	if report.Buckets["警觉"] > report.Buckets["舒缓"] {
		result = append(result,
			"警觉类情绪多于舒缓类情绪，可以回看这些日期前一晚的睡眠和现实压力。")
	} else if report.Buckets["舒缓"] > 0 {
		result = append(result,
			"窗口内存在稳定的舒缓类记录，建议保留当时的睡眠环境和睡前活动描述。")
	}
	if len(result) == 0 {
		result = append(result,
			"记录更多日期后，情绪转移和峰值会更有解释力。")
	}
	return result
}

func formatPercent(
	value float64,
) string {
	return fmt.Sprintf("%.0f%%", value*100)
}

func MoodCalendar(
	items []domain.Dream,
) map[string]MoodObservation {
	result := map[string]MoodObservation{}
	for i := range items {
		d := items[i]
		point := MoodObservation{Date: d.DreamDate.Format("2006-01-02"), Emotion: string(d.Emotion), Score: MoodScore(d.Emotion, d.Clarity), Bucket: text.MoodBucket(d.Emotion), Clarity: d.Clarity, SleepHours: d.SleepHours}
		if old, ok := result[point.Date]; !ok || point.Score > old.Score {
			result[point.Date] = point
		}
	}
	return result
}
