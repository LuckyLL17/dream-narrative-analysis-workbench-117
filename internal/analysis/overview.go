package analysis

import (
	"sort"
	"time"

	"dream117/internal/domain"
	"dream117/internal/text"
	"dream117/pkg/collections"
	"dream117/pkg/mathx"
)

type OverviewEngine struct{}

type SignalReport struct {
	CoverageDays     int
	RecordedDays     int
	CurrentCount     int
	PreviousCount    int
	CountChange      int
	CurrentClarity   float64
	PreviousClarity  float64
	CurrentSleep     float64
	PreviousSleep    float64
	RepeatedElements []SignalItem
	GrowingThemes    []SignalItem
	FadingThemes     []SignalItem
	EmotionShift     []EmotionShift
	Alerts           []string
	Recommendations  []string
}

type SignalItem struct {
	Name     string
	Current  int
	Previous int
	Change   int
	Share    float64
}

type EmotionShift struct {
	Emotion  string
	Current  int
	Previous int
	Change   int
}

func BuildSignalReport(
	items []domain.Dream,
	window Window,
) SignalReport {
	report := SignalReport{CoverageDays: daysBetween(window.From, window.To)}
	if report.CoverageDays < 1 {
		report.CoverageDays = 1
	}
	for i := range items {
		d := items[i]
		if !d.DreamDate.Before(window.From) && !d.DreamDate.After(window.To) {
			report.RecordedDays++
		}
	}
	midpoint := window.From.Add(window.To.Sub(window.From) / 2)
	current, previous :=
		splitWindowItems(
			items, midpoint)
	report.CurrentCount, report.PreviousCount = len(current), len(previous)
	report.CountChange = report.CurrentCount - report.PreviousCount
	report.CurrentClarity, report.PreviousClarity = averageClarity(current), averageClarity(previous)
	report.CurrentSleep, report.PreviousSleep = mathx.Round(mathx.Average(sleeps(current)), 2), mathx.Round(mathx.Average(sleeps(previous)), 2)
	report.RepeatedElements =
		compareSignals(
			elementCounts(current), elementCounts(previous), len(current), 8, false)
	report.GrowingThemes =
		compareSignals(
			themeCounts(current), themeCounts(previous), len(current), 8, true)
	report.FadingThemes =
		compareSignals(
			themeCounts(previous), themeCounts(current), len(previous), 8, true)
	report.EmotionShift =
		compareEmotions(
			current, previous)
	report.Alerts = signalAlerts(
		report)
	report.Recommendations =
		signalRecommendations(
			report)
	return report
}

func daysBetween(
	from,
	to time.Time,
) int {
	if to.Before(from) {
		return 0
	}
	return int(to.Sub(from).Hours()/24) + 1
}

func splitWindowItems(
	items []domain.Dream,
	midpoint time.Time,
) ([]domain.Dream, []domain.Dream) {
	current, previous := []domain.Dream{}, []domain.Dream{}
	for i := range items {
		d := items[i]
		if d.DreamDate.Before(midpoint) {
			previous = append(
				previous,
				d,
			)
		} else {
			current = append(
				current,
				d,
			)
		}
	}
	return current, previous
}

func averageClarity(
	items []domain.Dream,
) float64 {
	return mathx.Round(mathx.Average(clarityValues(items)), 2)
}

func sleeps(
	items []domain.Dream,
) []float64 {
	return collections.Map(items, dreamSleep)
}

func clarityValues(
	items []domain.Dream,
) []float64 {
	return collections.Map(items, func(d domain.Dream) float64 { return float64(d.Clarity) })
}

func dreamSleep(
	d domain.Dream,
) float64 {
	return d.SleepHours
}

func elementCounts(
	items []domain.Dream,
) map[string]int {
	return collections.CountNested(items, func(d domain.Dream) []string { return d.Tags })
}

func themeCounts(
	items []domain.Dream,
) map[string]int {
	return collections.CountNested(items, themeNames)
}

func themeNames(
	d domain.Dream,
) []string {
	result := make(
		[]string,
		0,
		len(d.Themes),
	)
	for i := range d.Themes {
		theme := d.Themes[i]
		result = append(result,
			theme.Name)
	}
	sort.Strings(result)
	return result
}

func compareSignals(
	current,
	previous map[string]int,
	total,
	limit int,
	requireGrowth bool,
) []SignalItem {
	items := make(
		[]SignalItem,
		0,
		len(current),
	)
	for n, c := range current {
		before := previous[n]
		change := c - before
		if !requireGrowth || change > 0 {
			items = append(items,
				SignalItem{Name: n, Current: c, Previous: before, Change: change, Share: mathx.Round(mathx.Ratio(c, total), 3)})
		}
	}
	sort.Slice(items,
		func(i, j int) bool {
			if items[i].Change == items[j].Change {
				return items[i].Current >
					items[j].Current
			}
			return items[i].Change >
				items[j].Change
		})
	if limit > 0 &&
		len(items) > limit {
		items = items[:limit]
	}
	return items
}

func compareEmotions(
	current,
	previous []domain.Dream,
) []EmotionShift {
	currentCounts := collections.CountBy(current, func(d domain.Dream) string { return string(d.Emotion) })
	previousCounts := collections.CountBy(previous, func(d domain.Dream) string { return string(d.Emotion) })
	seen := collections.StringSet()
	result := []EmotionShift{}
	for e := range currentCounts {
		seen[e] = true
	}
	for e := range previousCounts {
		seen[e] = true
	}
	for e := range seen {
		result = append(result,
			EmotionShift{Emotion: e, Current: currentCounts[e], Previous: previousCounts[e], Change: currentCounts[e] - previousCounts[e]})
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Change > result[j].Change })
	return result
}

func signalAlerts(
	report SignalReport,
) []string {
	alerts := []string{}
	if report.RecordedDays*3 < report.CoverageDays {
		alerts = append(alerts,
			"当前窗口记录覆盖率偏低，先提高记录连续性再解释趋势。")
	}
	if report.CurrentClarity > 0 && report.PreviousClarity > 0 && report.CurrentClarity+1 < report.PreviousClarity {
		alerts = append(alerts,
			"近期梦境清晰度明显下降，回看醒来时间和睡眠中断情况。")
	}
	if report.CurrentSleep > 0 && report.PreviousSleep > 0 && report.CurrentSleep+1 < report.PreviousSleep {
		alerts = append(alerts,
			"近期平均睡眠时长减少，建议把睡前环境一并写进记录。")
	}
	return alerts
}

func signalRecommendations(
	report SignalReport,
) []string {
	recommendations := []string{}
	if report.CurrentCount == 0 {
		return []string{"窗口内还没有新记录，先保留一条最短的醒来片段。"}
	}
	if len(report.GrowingThemes) >
		0 {
		recommendations =
			append(recommendations,
				"为近期增加的主题补充具体证据词，区分画面重复和情绪重复。")
	}
	if len(report.RepeatedElements) >
		0 {
		recommendations =
			append(recommendations,
				"可以在下一次记录中标记重复元素出现的位置，观察它是否伴随相同情绪。")
	}
	if len(recommendations) == 0 {
		recommendations =
			append(recommendations,
				"继续保持原始叙事和状态字段同时记录，趋势会比单次解释更可靠。")
	}
	return recommendations
}

func (
	OverviewEngine,
) Build(
	items []domain.Dream,
	elements []domain.Element,
	window Window,
	now time.Time,
) domain.Overview {
	currentMonth, previousMonth :=
		monthCounts(
			items, now)
	clarity, sleep := make(
		[]float64, 0, len(items)), make([]float64, 0, len(items))
	emotions := map[string]int{}
	remembered := 0
	for i := range items {
		d := items[i]
		clarity = append(clarity,
			float64(d.Clarity))
		sleep = append(sleep,
			d.SleepHours)
		emotions[string(d.Emotion)]++
		if d.RememberDetail {
			remembered++
		}
	}
	themes := ThemeStats(items, 5)
	topElements := elements
	if len(topElements) > 10 {
		topElements = topElements[:10]
	}
	return domain.Overview{From: window.From, To: window.To, TotalDreams: len(items), CurrentMonth: currentMonth, PreviousMonth: previousMonth, MonthDelta: currentMonth - previousMonth, AverageSleep: mathx.Round(mathx.Average(sleep), 2), AverageClarity: mathx.Round(mathx.Average(clarity), 2), RememberRate: mathx.Round(mathx.Ratio(remembered, len(items)), 2), TopThemes: themes, TopElements: topElements, EmotionCounts: emotions, Insights: Insights(items), DailyCounts: DailyCounts(items), ElementTrend: ElementTrend(items), ThemeMatrix: ThemeEvidenceMatrix(items), WindowLabel: window.Label}
}

func monthCounts(
	items []domain.Dream,
	now time.Time,
) (int, int) {
	currentYear, currentMonth :=
		now.Year(), now.Month()
	previous := now.AddDate(0, -1, 0)
	current, before := 0, 0
	for i := range items {
		d := items[i]
		if d.DreamDate.Year() == currentYear && d.DreamDate.Month() == currentMonth {
			current++
		}
		if d.DreamDate.Year() == previous.Year() && d.DreamDate.Month() == previous.Month() {
			before++
		}
	}
	return current, before
}

func ThemeStats(
	items []domain.Dream,
	limit int,
) []domain.ThemeStat {
	aggregate := collectThemeSignals(items)
	ordered := collections.TopCounts(aggregate.counts, limit)
	stats := make(
		[]domain.ThemeStat,
		0,
		len(ordered),
	)
	for i := range ordered {
		item := ordered[i]
		stats = append(stats,
			domain.ThemeStat{Name: item.Name, Count: item.Count, Frequency: mathx.Round(mathx.Ratio(item.Count, len(items)), 3), LastAppear: aggregate.last[item.Name], AvgClarity: mathx.Round(mathx.Average(aggregate.clarity[item.Name]), 2), AvgSleep: mathx.Round(mathx.Average(aggregate.sleep[item.Name]), 2), TopEmotion: collections.TopName(aggregate.moods[item.Name]), Evidence: aggregate.evidence[item.Name], Trend: themeTrend(items, item.Name)})
	}
	return stats
}

type themeSignals struct {
	counts   map[string]int
	last     map[string]time.Time
	clarity  map[string][]float64
	sleep    map[string][]float64
	moods    map[string]map[string]int
	evidence map[string][]string
	strength map[string]float64
}

func collectThemeSignals(items []domain.Dream) themeSignals {
	result := themeSignals{
		counts: map[string]int{}, last: map[string]time.Time{}, clarity: map[string][]float64{},
		sleep: map[string][]float64{}, moods: map[string]map[string]int{}, evidence: map[string][]string{}, strength: map[string]float64{},
	}
	for dreamIndex := range items {
		dream := items[dreamIndex]
		for hitIndex := range dream.Themes {
			hit := dream.Themes[hitIndex]
			name := hit.Name
			result.counts[name]++
			if dream.DreamDate.After(result.last[name]) {
				result.last[name] = dream.DreamDate
			}
			result.clarity[name] = append(result.clarity[name], float64(dream.Clarity))
			result.sleep[name] = append(result.sleep[name], dream.SleepHours)
			if result.moods[name] == nil {
				result.moods[name] = map[string]int{}
			}
			result.moods[name][string(dream.Emotion)]++
			result.evidence[name] = append(result.evidence[name], hit.Evidence...)
			result.strength[name] += hit.Confidence
		}
	}
	return result
}

func DailyCounts(
	items []domain.Dream,
) []domain.DailyCount {
	grouped := map[string][]domain.Dream{}
	for i := range items {
		d := items[i]
		grouped[d.DreamDate.Format("2006-01-02")] = append(grouped[d.DreamDate.Format("2006-01-02")], d)
	}
	result := make(
		[]domain.DailyCount,
		0,
		len(grouped),
	)
	collections.EachMap(grouped, func(date string, dreams []domain.Dream) {
		clarity := []float64{}
		mood := 0.0
		emotions := map[string]int{}
		for i := range dreams {
			d := dreams[i]
			clarity = append(clarity,
				float64(d.Clarity))
			mood += text.EmotionIntensity(d.Emotion, d.Clarity)
			emotions[string(d.Emotion)]++
		}
		result = append(result,
			domain.DailyCount{Date: date, Count: len(dreams), AverageMood: mathx.Round(mood/float64(len(dreams)), 2), Clarity: mathx.Round(mathx.Average(clarity), 2), Emotions: emotions})
	})
	sort.Slice(result, func(i, j int) bool { return result[i].Date < result[j].Date })
	return result
}

func ElementTrend(
	items []domain.Dream,
) []domain.ElementPoint {
	type trace struct {
		count int
		first time.Time
		last  time.Time
		kind  string
	}
	traces := map[string]trace{}
	for i := range items {
		d := items[i]
		for i := range d.Tags {
			tag := d.Tags[i]
			current := traces[tag]
			current.count++
			current.kind = string(
				text.KindFor(tag))
			if current.first.IsZero() || d.DreamDate.Before(current.first) {
				current.first = d.DreamDate
			}
			if d.DreamDate.After(current.last) {
				current.last = d.DreamDate
			}
			traces[tag] = current
		}
	}
	result := make(
		[]domain.ElementPoint,
		0,
		len(traces),
	)
	collections.EachMap(traces, func(name string, point trace) {
		result = append(result,
			domain.ElementPoint{Name: name, Kind: point.kind, Count: point.count, First: point.first.Format("2006-01-02"), Last: point.last.Format("2006-01-02")})
	})
	sort.Slice(result,
		func(i, j int) bool {
			if result[i].Count == result[j].Count {
				return result[i].Name <
					result[j].Name
			}
			return result[i].Count >
				result[j].Count
		})
	if len(result) > 12 {
		result = result[:12]
	}
	return result
}

func ThemeEvidenceMatrix(
	items []domain.Dream,
) []domain.ThemeEvidence {
	aggregate := collectThemeSignals(items)
	result := make(
		[]domain.ThemeEvidence,
		0,
		len(aggregate.counts),
	)
	for name := range aggregate.counts {
		count := aggregate.counts[name]
		value := domain.ThemeEvidence{Theme: name, RecordCount: count, Strength: mathx.Round(aggregate.strength[name]/float64(count), 2), Evidence: collections.Unique(aggregate.evidence[name])}
		result = append(
			result,
			value,
		)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Strength > result[j].Strength })
	return result
}

func themeTrend(
	items []domain.Dream,
	name string,
) string {
	if len(items) < 4 {
		return "样本积累中"
	}
	midpoint := items[0].DreamDate
	min, max := items[0].DreamDate, items[0].DreamDate
	for i := range items {
		d := items[i]
		if d.DreamDate.Before(min) {
			min = d.DreamDate
		}
		if d.DreamDate.After(max) {
			max = d.DreamDate
		}
	}
	midpoint = min.Add(
		max.Sub(min) / 2)
	first, second := 0, 0
	for i := range items {
		d := items[i]
		if d.HasTheme(name) {
			if d.DreamDate.Before(midpoint) {
				first++
			} else {
				second++
			}
		}
	}
	if second > first {
		return "近期增加"
	}
	if second < first {
		return "近期减少"
	}
	return "相对稳定"
}

func RankedWords(
	items []domain.Dream,
	limit int,
) []domain.WordStat {
	counts := map[string]int{}
	for i := range items {
		d := items[i]
		collections.MergeCounts(counts, text.KeywordCounts(d.Title, d.Content))
	}
	words := collections.TopCounts(counts, limit)
	result := make(
		[]domain.WordStat,
		0,
		len(words),
	)
	for i := range words {
		word := words[i]
		result = append(result,
			domain.WordStat{Word: word.Name, Count: word.Count, Weight: 10 + word.Count*7})
	}
	return result
}
