package analysis

import (
	"fmt"
	"sort"
	"strconv"

	"dream117/internal/domain"
	"dream117/internal/text"
	"dream117/pkg/collections"
	"dream117/pkg/mathx"
)

type PatternReport struct {
	Weekdays       []WeekdayPattern
	Months         []MonthPattern
	Quality        QualitySummary
	ThemeRhythms   []ThemeRhythm
	WritingSignals WritingSignals
	Stability      StabilityProfile
	Summary        []string
}

type StabilityProfile struct {
	SleepRange     float64
	ClarityRange   float64
	EmotionVariety int
	StableSleep    bool
	StableClarity  bool
	Interpretation string
}

type WeekdayPattern struct {
	Weekday      int
	Name         string
	Count        int
	Share        float64
	AverageMood  float64
	AverageSleep float64
	TopThemes    []string
}

type MonthPattern struct {
	Month          string
	Count          int
	AverageClarity float64
	AverageSleep   float64
	RememberRate   float64
	TopEmotion     string
}

type QualitySummary struct {
	Complete      int
	Partial       int
	WithThemes    int
	WithElements  int
	AverageLength float64
	MissingFields []string
}

type ThemeRhythm struct {
	Theme       string
	ByWeekday   map[string]int
	PeakWeekday string
	PeakCount   int
}

type WritingSignals struct {
	AverageSentences float64
	AverageActions   float64
	AveragePlaces    float64
	AverageSensory   float64
	QuestionRate     float64
	DenseRecords     int
}

func BuildPatternReport(
	items []domain.Dream,
) PatternReport {
	report := PatternReport{Weekdays: weekdayPatterns(items), Months: monthPatterns(items), Quality: qualitySummary(items), ThemeRhythms: themeRhythms(items), WritingSignals: writingSignals(items), Stability: stabilityProfile(items)}
	report.Summary =
		patternSummary(
			report)
	return report
}

func stabilityProfile(
	items []domain.Dream,
) StabilityProfile {
	result := StabilityProfile{}
	if len(items) == 0 {
		result.Interpretation = "还没有足够记录判断状态稳定性。"
		return result
	}
	emotions := collections.CountBy(items, func(d domain.Dream) string { return string(d.Emotion) })
	sleeps, clarities := sleeps(
		items), clarityValues(items)
	result.SleepRange =
		mathx.Round(
			valueRange(sleeps), 2)
	result.ClarityRange =
		mathx.Round(
			valueRange(clarities), 2)
	result.EmotionVariety = len(
		emotions)
	result.StableSleep = result.SleepRange <= 2
	result.StableClarity = result.ClarityRange <= 3
	switch {
	case result.StableSleep && result.StableClarity:
		result.Interpretation = "睡眠和清晰度都相对集中，适合继续观察主题变化。"
	case !result.StableSleep && !result.StableClarity:
		result.Interpretation = "睡眠和清晰度波动都比较明显，单次梦境不宜代表整体趋势。"
	case !result.StableSleep:
		result.Interpretation = "睡眠时长变化较大，可以把主题和睡眠分组对照。"
	default:
		result.Interpretation = "清晰度变化较大，回看时要区分记忆强弱和主题频率。"
	}
	return result
}

func valueRange(
	values []float64,
) float64 {
	if len(values) == 0 {
		return 0
	}
	min, max := values[0], values[0]
	for i := range values[1:] {
		value := values[1:][i]
		if value < min {
			min = value
		}
		if value > max {
			max = value
		}
	}
	return max - min
}

func weekdayPatterns(
	items []domain.Dream,
) []WeekdayPattern {
	type bucket struct {
		count  int
		mood   []float64
		sleep  []float64
		themes map[string]int
	}
	buckets := make([]bucket, 7)
	for i := range items {
		d := items[i]
		day := int(d.DreamDate.Weekday())
		buckets[day].count++
		buckets[day].mood = append(buckets[day].mood, text.EmotionIntensity(d.Emotion, d.Clarity))
		buckets[day].sleep = append(buckets[day].sleep, d.SleepHours)
		if buckets[day].themes == nil {
			buckets[day].themes = map[string]int{}
		}
		for i := range d.Themes {
			theme := d.Themes[i]
			buckets[day].themes[theme.Name]++
		}
	}
	result := make(
		[]WeekdayPattern, 0, 7)
	for day := range buckets {
		bucket := buckets[day]
		result = append(result,
			WeekdayPattern{Weekday: day, Name: weekdayName(day), Count: bucket.count, Share: mathx.Round(mathx.Ratio(bucket.count, len(items)), 3), AverageMood: mathx.Round(mathx.Average(bucket.mood), 3), AverageSleep: mathx.Round(mathx.Average(bucket.sleep), 2), TopThemes: topThemeNames(bucket.themes, 3)})
	}
	return result
}

func weekdayName(
	value int,
) string {
	return []string{"周日", "周一", "周二", "周三", "周四", "周五", "周六"}[value%7]
}

func monthPatterns(
	items []domain.Dream,
) []MonthPattern {
	type bucket struct {
		clarity  []float64
		sleep    []float64
		remember int
		emotions map[string]int
		count    int
	}
	buckets := map[string]bucket{}
	for i := range items {
		d := items[i]
		key := d.DreamDate.Format("2006-01")
		current := buckets[key]
		current.count++
		current.clarity =
			append(current.clarity,
				float64(d.Clarity))
		current.sleep =
			append(current.sleep,
				d.SleepHours)
		if d.RememberDetail {
			current.remember++
		}
		if current.emotions == nil {
			current.emotions = map[string]int{}
		}
		current.emotions[string(d.Emotion)]++
		buckets[key] = current
	}
	months := collections.Keys(buckets)
	result := make(
		[]MonthPattern,
		0,
		len(months),
	)
	for i := range months {
		month := months[i]
		bucket := buckets[month]
		result = append(result,
			MonthPattern{Month: month, Count: bucket.count, AverageClarity: mathx.Round(mathx.Average(bucket.clarity), 2), AverageSleep: mathx.Round(mathx.Average(bucket.sleep), 2), RememberRate: mathx.Round(mathx.Ratio(bucket.remember, bucket.count), 3), TopEmotion: collections.TopName(bucket.emotions)})
	}
	return result
}

func qualitySummary(
	items []domain.Dream,
) QualitySummary {
	result := QualitySummary{}
	lengths := []float64{}
	missing := collections.Counter()
	for i := range items {
		d := items[i]
		lengths = append(lengths,
			float64(len([]rune(d.Content))))
		complete := true
		if d.Title == "" {
			missing["标题"]++
			complete = false
		}
		if d.Content == "" {
			missing["正文"]++
			complete = false
		}
		if d.SleepHours <= 0 {
			missing["睡眠时长"]++
			complete = false
		}
		if d.Emotion == "" {
			missing["情绪"]++
			complete = false
		}
		if complete {
			result.Complete++
		} else {
			result.Partial++
		}
		if len(d.Themes) > 0 {
			result.WithThemes++
		}
		if len(d.Tags) > 0 {
			result.WithElements++
		}
	}
	result.AverageLength =
		mathx.Round(
			mathx.Average(lengths), 1)
	for n, c := range missing {
		if c > 0 {
			result.MissingFields =
				append(result.MissingFields,
					n)
		}
	}
	sort.Strings(result.MissingFields)
	return result
}

func themeRhythms(
	items []domain.Dream,
) []ThemeRhythm {
	groups := map[string]map[string]int{}
	for i := range items {
		d := items[i]
		weekday := weekdayName(int(d.DreamDate.Weekday()))
		for i := range d.Themes {
			theme := d.Themes[i]
			if groups[theme.Name] == nil {
				groups[theme.Name] = map[string]int{}
			}
			groups[theme.Name][weekday]++
		}
	}
	result := make(
		[]ThemeRhythm,
		0,
		len(groups),
	)
	collections.EachMap(groups, func(theme string, counts map[string]int) {
		peak, count := "", 0
		collections.EachMap(counts, func(day string, value int) {
			if value > count {
				peak, count = day, value
			}
		})
		result = append(result,
			ThemeRhythm{Theme: theme, ByWeekday: counts, PeakWeekday: peak, PeakCount: count})
	})
	sort.Slice(result, func(i, j int) bool { return result[i].PeakCount > result[j].PeakCount })
	if len(result) > 12 {
		result = result[:12]
	}
	return result
}

func writingSignals(
	items []domain.Dream,
) WritingSignals {
	result := WritingSignals{}
	if len(items) == 0 {
		return result
	}
	for i := range items {
		d := items[i]
		profile := text.Profile(d.Title, d.Content)
		result.AverageSentences += float64(profile.Sentences)
		result.AverageActions += float64(profile.Actions)
		result.AveragePlaces += float64(profile.Places)
		result.AverageSensory += float64(profile.SensoryWords)
		if profile.Questions > 0 {
			result.QuestionRate++
		}
		if profile.Density >= 8 {
			result.DenseRecords++
		}
	}
	result.AverageSentences =
		mathx.Round(
			result.AverageSentences/float64(len(items)), 2)
	result.AverageActions =
		mathx.Round(
			result.AverageActions/float64(len(items)), 2)
	result.AveragePlaces =
		mathx.Round(
			result.AveragePlaces/float64(len(items)), 2)
	result.AverageSensory =
		mathx.Round(
			result.AverageSensory/float64(len(items)), 2)
	result.QuestionRate =
		mathx.Round(
			result.QuestionRate/float64(len(items)), 3)
	return result
}

func topThemeNames(
	counts map[string]int,
	limit int,
) []string {
	return collections.Names(collections.TopCounts(counts, limit))
}

func patternSummary(
	report PatternReport,
) []string {
	result := []string{}
	bestDay := WeekdayPattern{}
	for i := range report.Weekdays {
		day := report.Weekdays[i]
		if day.Count > bestDay.Count {
			bestDay = day
		}
	}
	if bestDay.Count > 0 {
		result = append(result,
			bestDay.Name+"是目前记录最多的日期，共有 "+itoa(bestDay.Count)+" 条。")
	}
	if report.Quality.AverageLength > 0 {
		result = append(result,
			"每条记录平均 "+itoa(int(report.Quality.AverageLength))+" 个字，叙事密度可以结合清晰度一起观察。")
	}
	if report.WritingSignals.QuestionRate > 0.4 {
		result = append(result,
			"超过四成记录包含疑问句，醒来时的未完成感可能值得单独标记。")
	}
	if len(result) == 0 {
		result = append(result,
			"继续积累跨日期记录，星期、月份和叙事密度的差异会逐渐稳定。")
	}
	return result
}

func itoa(
	value int,
) string {
	return strconv.Itoa(value)
}

type ReviewDigest struct {
	Window      string
	Headline    string
	Subheadline string
	Confidence  string
	Facts       []ReviewFact
	Focus       []ReviewQuestion
	Warnings    []string
	Motifs      []ReviewMotif
	Quality     ReviewQuality
	NextWindow  []string
}

type ReviewFact struct {
	Label string
	Value string
	Note  string
}

type ReviewQuestion struct {
	Question string
	Reason   string
	Source   string
}

type ReviewMotif struct {
	Name      string
	Count     int
	Strength  float64
	Evidence  []string
	Direction string
}

type ReviewQuality struct {
	RecordedDays int
	Coverage     int
	Completion   float64
	Missing      []string
}

func BuildReviewDigest(
	items []domain.Dream,
	window Window,
) ReviewDigest {
	mood := BuildMoodReport(items)
	signals := BuildSignalReport(items, window)
	patterns := BuildPatternReport(items)
	facets := BuildFacets(items)
	quality := ReviewQuality{RecordedDays: signals.RecordedDays, Coverage: signals.CoverageDays, Completion: ratioReview(signals.RecordedDays, signals.CoverageDays), Missing: clarityMissing(facets)}
	digest := ReviewDigest{Window: window.Label, Quality: quality}
	digest.Headline =
		reviewHeadline(
			items, mood, signals)
	digest.Subheadline =
		reviewSubheadline(
			items, patterns, facets)
	digest.Confidence =
		reviewConfidence(
			quality, len(items))
	digest.Facts = reviewFacts(
		items, mood, signals, facets)
	digest.Motifs = reviewMotifs(
		items, signals)
	digest.Focus =
		reviewQuestions(
			items, mood, patterns, facets)
	digest.Warnings =
		reviewWarnings(
			signals, facets)
	digest.NextWindow =
		reviewNextWindow(
			digest)
	return digest
}

func clarityMissing(
	report FacetReport,
) []string {
	result := []string{}
	for i := range report.Clarity {
		value := report.Clarity[i]
		if value.Count == 0 {
			result = append(result,
				value.Label)
		}
	}
	return result
}

func ratioReview(
	value,
	total int,
) float64 {
	if total <= 0 {
		return 0
	}
	return float64(value) / float64(total)
}

func reviewHeadline(
	items []domain.Dream,
	mood MoodReport,
	signals SignalReport,
) string {
	if len(items) == 0 {
		return "这一段时间还没有留下梦境"
	}
	if len(signals.GrowingThemes) >
		0 {
		return "近期有一条主题正在变得更响"
	}
	if mood.Dominant != "" {
		return "最近的醒来情绪以“" + mood.Dominant + "”为主"
	}
	return "你的梦境记录正在形成自己的节奏"
}

func reviewSubheadline(
	items []domain.Dream,
	patterns PatternReport,
	facets FacetReport,
) string {
	if len(items) == 1 {
		return "先把这一次记完整，再等待下一次相似画面出现。"
	}
	if len(patterns.ThemeRhythms) >
		0 {
		return fmt.Sprintf("已经记录 %d 条，主题线索在日期和叙事密度中逐渐有了轮廓。", len(items))
	}
	if len(facets.Elements) >
		0 {
		return "元素已经积累起来，继续保留原始用词会更方便后续比较。"
	}
	return "继续同时记录画面、情绪和睡眠，避免只留下结论。"
}

func reviewConfidence(
	quality ReviewQuality,
	total int,
) string {
	if total < 3 || quality.Completion < 0.15 {
		return "样本较少"
	}
	if total < 8 || quality.Completion < 0.35 {
		return "初步观察"
	}
	return "可持续观察"
}

func reviewFacts(
	items []domain.Dream,
	mood MoodReport,
	signals SignalReport,
	facets FacetReport,
) []ReviewFact {
	facts := []ReviewFact{{Label: "梦境记录", Value: fmt.Sprintf("%d 条", len(items)), Note: "只统计当前分析窗口内的记录。"}, {Label: "记录覆盖", Value: fmt.Sprintf("%d/%d 天", signals.RecordedDays, signals.CoverageDays), Note: "覆盖率越稳定，变化信号越容易解释。"}}
	if mood.Dominant != "" {
		facts = append(facts,
			ReviewFact{Label: "主导情绪", Value: mood.Dominant, Note: fmt.Sprintf("占比约 %.0f%%。", mood.DominantShare*100)})
	}
	if len(facets.Themes) > 0 {
		facts = append(facts,
			ReviewFact{Label: "高频主题", Value: facets.Themes[0].Label, Note: fmt.Sprintf("出现 %d 次。", facets.Themes[0].Count)})
	}
	if signals.CurrentClarity > 0 {
		facts = append(facts,
			ReviewFact{Label: "近期清晰度", Value: fmt.Sprintf("%.1f/10", signals.CurrentClarity), Note: compareNumber(signals.CurrentClarity, signals.PreviousClarity, "与前半段相比")})
	}
	if signals.CurrentSleep > 0 {
		facts = append(facts,
			ReviewFact{Label: "近期睡眠", Value: fmt.Sprintf("%.1f 小时", signals.CurrentSleep), Note: compareNumber(signals.CurrentSleep, signals.PreviousSleep, "与前半段相比")})
	}
	return facts
}

func compareNumber(
	current,
	previous float64,
	prefix string,
) string {
	if previous <= 0 {
		return "前半段样本不足。"
	}
	delta := current - previous
	if delta > 0.2 {
		return prefix + "有所增加。"
	}
	if delta < -0.2 {
		return prefix + "有所下降。"
	}
	return prefix + "基本稳定。"
}

func reviewMotifs(
	items []domain.Dream,
	signals SignalReport,
) []ReviewMotif {
	counts, strength, evidence := collections.Counter(), map[string]float64{}, map[string][]string{}
	for i := range items {
		d := items[i]
		for i := range d.Themes {
			theme := d.Themes[i]
			counts[theme.Name]++
			strength[theme.Name] += theme.Confidence
			evidence[theme.Name] = append(evidence[theme.Name], theme.Evidence...)
		}
	}
	result := []ReviewMotif{}
	collections.EachCount(counts, func(name string, count int) {
		direction := "稳定"
		for i := range signals.GrowingThemes {
			growing := signals.GrowingThemes[i]
			if growing.Name == name {
				direction = "增加"
			}
		}
		for i := range signals.FadingThemes {
			fading := signals.FadingThemes[i]
			if fading.Name == name {
				direction = "减少"
			}
		}
		result = append(result,
			ReviewMotif{Name: name, Count: count, Strength: strength[name] / float64(count), Evidence: collections.Unique(evidence[name]), Direction: direction})
	})
	sort.Slice(result, func(i, j int) bool { return result[i].Count > result[j].Count })
	if len(result) > 8 {
		result = result[:8]
	}
	return result
}

func reviewQuestions(
	items []domain.Dream,
	mood MoodReport,
	patterns PatternReport,
	facets FacetReport,
) []ReviewQuestion {
	result := []ReviewQuestion{}
	if len(mood.Transitions) >
		0 {
		result = append(result,
			ReviewQuestion{Question: "哪一种情绪转移最常发生？", Reason: "情绪序列比单个标签更能说明醒来后的变化。", Source: "情绪报告"})
	}
	if len(patterns.ThemeRhythms) >
		0 {
		result = append(result,
			ReviewQuestion{Question: "主题是否集中在某几个星期几？", Reason: "日期节律可以帮助你回看固定的作息和现实安排。", Source: "星期模式"})
	}
	if facets.WithDetail < len(items) {
		result = append(result,
			ReviewQuestion{Question: "哪些记录只有情绪，没有留下具体画面？", Reason: "细节记忆率会影响主题证据的可复核性。", Source: "记录质量"})
	}
	if len(result) == 0 {
		result = append(result,
			ReviewQuestion{Question: "下一次醒来时，哪个细节最值得先写下来？", Reason: "先保留原始记忆，再做解释。", Source: "基础回看"})
	}
	return result
}

func reviewWarnings(
	signals SignalReport,
	facets FacetReport,
) []string {
	result := append([]string(nil), signals.Alerts...)
	if facets.Total > 0 && facets.WithDetail*2 < facets.Total {
		result = append(result,
			"超过一半记录没有标记为记得具体细节，回看时要区分完整记忆和片段印象。")
	}
	return collections.Unique(result)
}

func reviewNextWindow(
	digest ReviewDigest,
) []string {
	result := []string{}
	if len(digest.Warnings) >
		0 {
		result = append(result,
			"优先补齐记录质量，再比较主题变化。")
	}
	if len(digest.Motifs) > 0 {
		result = append(result,
			"保留最高频主题的原始证据词，观察它是否在下一窗口继续出现。")
	}
	result = append(result,
		"继续记录醒来情绪、清晰度和睡眠时长，避免只记录内容。")
	return collections.Unique(result)
}
