package analysis

import (
	"fmt"
	"strings"
	"time"

	"dream117/internal/domain"
	"dream117/pkg/collections"
	"dream117/pkg/mathx"
)

func BuildReport(
	userID string,
	items []domain.Dream,
	start,
	end time.Time,
	now time.Time,
) domain.WeeklyReport {
	stats := ThemeStats(items, 3)
	emotion := EmotionCounts(items)
	clarity, sleep := []float64{}, []float64{}
	for i := range items {
		d := items[i]
		clarity = append(clarity,
			float64(d.Clarity))
		sleep = append(sleep,
			d.SleepHours)
	}
	state := domain.ReportReady
	summary := reportSummary(items, stats, emotion)
	suggestions := reportSuggestions(items, stats)
	if len(items) == 0 {
		state = domain.ReportPending
		summary = "这一周还没有梦境记录，醒来后留下一句也算一次观察。"
	}
	return domain.WeeklyReport{ID: fmt.Sprintf("weekly-%s", start.Format("20060102")), UserID: userID, WeekStart: start, WeekEnd: end, State: state, TotalDreams: len(items), TopThemes: stats, EmotionDist: emotion, AvgClarity: mathx.Round(mathx.Average(clarity), 2), AvgSleep: mathx.Round(mathx.Average(sleep), 2), Summary: summary, Suggestions: suggestions, CreatedAt: now, UpdatedAt: now}
}

func reportSummary(
	items []domain.Dream,
	themes []domain.ThemeStat,
	emotions map[string]int,
) string {
	if len(items) == 0 {
		return ""
	}
	mood := ""
	max := 0
	for k := range emotions {
		n := emotions[k]
		if n > max {
			mood, max = k, n
		}
	}
	if len(themes) == 0 {
		return fmt.Sprintf("本周记录 %d 条，醒来时最常见的情绪是“%s”，主题线索还在积累。", len(items), mood)
	}
	return fmt.Sprintf("本周记录 %d 条，最常见的醒来情绪是“%s”，重复线索集中在“%s”。", len(items), mood, themes[0].Name)
}

func reportSuggestions(
	items []domain.Dream,
	themes []domain.ThemeStat,
) []string {
	result := []string{}
	if len(items) < 3 {
		result = append(result,
			"保持连续记录，至少积累三条后再比较主题变化。")
	}
	if len(themes) > 0 {
		result = append(result,
			"回看“"+themes[0].Name+"”出现前一天的睡眠与情绪，寻找现实情境中的对应线索。")
	}
	for i := range items {
		d := items[i]
		if d.SleepHours < 6 {
			result = append(result,
				"本周出现少于 6 小时睡眠的记录，下一周可以单独观察清晰度和害怕/紧张情绪。")
			break
		}
	}
	if len(result) == 0 {
		result = append(result,
			"下周可以在正文中补充一个具体地点或颜色，让叙事更容易被系统识别。")
	}
	return uniqueStrings(result)
}

func uniqueStrings(
	values []string,
) []string {
	result := make([]string, 0, len(values))
	for index := range values {
		v := values[index]
		if strings.TrimSpace(v) != "" {
			result = append(result, v)
		}
	}
	return collections.Unique(result)
}
