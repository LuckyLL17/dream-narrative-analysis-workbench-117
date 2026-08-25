package analysis

import (
	"fmt"

	"dream117/internal/domain"
	"dream117/internal/text"
	"dream117/pkg/collections"
	"dream117/pkg/mathx"
)

func Insights(
	items []domain.Dream,
) []domain.Insight {
	if len(items) == 0 {
		return []domain.Insight{{Title: "先留下一小段记忆", Detail: "记录标题、醒来情绪和一个画面，系统才能逐步找到你的个人线索。", Confidence: 1, Tone: "warm"}}
	}
	result := make(
		[]domain.Insight, 0, 4)
	result = append(result,
		sleepInsight(items))
	result = append(result,
		clarityInsight(items))
	result = append(result,
		themeInsight(items))
	result = append(result,
		moodInsight(items))
	collections.SortBy(
		result,
		func(left, right domain.Insight) bool {
			return left.Confidence > right.Confidence
		},
	)
	return result
}

func sleepInsight(
	items []domain.Dream,
) domain.Insight {
	short, long := []float64{}, []float64{}
	for i := range items {
		d := items[i]
		if d.SleepHours < 6 {
			short = append(short,
				float64(d.Clarity))
		}
		if d.SleepHours >= 8 {
			long = append(long,
				float64(d.Clarity))
		}
	}
	if len(short) == 0 || len(long) == 0 {
		return domain.Insight{Title: "继续积累睡眠线索", Detail: "目前短睡眠或充足睡眠样本还不够，保持记录后会更容易比较清晰度差异。", Confidence: 0.42, Tone: "neutral"}
	}
	difference := mathx.Average(long) - mathx.Average(short)
	return domain.Insight{Title: "睡眠时长与清晰度", Detail: fmt.Sprintf("睡眠至少 8 小时的记录，清晰度平均比少于 6 小时的记录高 %.1f 分。", difference), Confidence: mathx.Ratio(len(short)+len(long), len(items)*2), Tone: "violet"}
}

func clarityInsight(
	items []domain.Dream,
) domain.Insight {
	high := 0
	for i := range items {
		d := items[i]
		if d.Clarity >= 8 {
			high++
		}
	}
	return domain.Insight{Title: "细节保留率", Detail: fmt.Sprintf("%d/%d 条记录的清晰度达到 8 分以上，醒来后先写下画面再补充解释，有助于保留原始细节。", high, len(items)), Confidence: mathx.Ratio(len(items), len(items)+2), Tone: "blue"}
}

func themeInsight(
	items []domain.Dream,
) domain.Insight {
	stats := ThemeStats(items, 1)
	if len(stats) == 0 {
		return domain.Insight{Title: "主题还在形成", Detail: "你的记录暂时没有命中预设主题，可以用自然语言继续书写，系统会根据证据词更新线索。", Confidence: 0.38, Tone: "neutral"}
	}
	hit := stats[0]
	return domain.Insight{Title: "重复线索：" + hit.Name, Detail: fmt.Sprintf("该主题出现 %d 次，占当前窗口记录的 %.0f%%。证据不代表固定含义，建议结合当时的现实情境回看。", hit.Count, hit.Frequency*100), Confidence: min(0.98, 0.5+hit.Frequency), Tone: "amber"}
}

func moodInsight(
	items []domain.Dream,
) domain.Insight {
	counts := EmotionCounts(items)
	best := ""
	max := 0
	for e, n := range counts {
		if n > max {
			max, best = n, e
		}
	}
	return domain.Insight{Title: "醒来情绪主色", Detail: fmt.Sprintf("当前记录中“%s”出现最多，共 %d 次。情绪标签是自我观察入口，不是诊断结论。", best, max), Confidence: mathx.Ratio(max, len(items)), Tone: "rose"}
}

func min(
	a,
	b float64,
) float64 {
	if a < b {
		return a
	}
	return b
}

var _ = text.MoodBucket
