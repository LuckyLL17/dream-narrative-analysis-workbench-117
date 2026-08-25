package analysis

import (
	"sort"
	"strings"

	"dream117/internal/domain"
	"dream117/internal/text"
	"dream117/pkg/collections"
)

type ReflectionReport struct {
	DreamID      string
	Title        string
	Date         string
	Completeness float64
	Profile      text.NarrativeProfile
	Prompts      []ReflectionPrompt
	Missing      []string
	Evidence     []ReflectionEvidence
	NextActions  []string
	SimilarTerms []string
}

type ReflectionPrompt struct {
	Title    string
	Question string
	Reason   string
	Priority string
}

type ReflectionEvidence struct {
	Kind  string
	Name  string
	Words []string
	Note  string
}

func BuildReflection(
	d domain.Dream,
) ReflectionReport {
	report := ReflectionReport{DreamID: d.ID, Title: d.Title, Date: d.DreamDate.Format("2006-01-02"), Profile: text.Profile(d.Title, d.Content)}
	report.Missing =
		missingFields(
			d)
	report.Completeness =
		completeness(
			d, len(report.Missing))
	report.Evidence =
		reflectionEvidence(
			d)
	report.Prompts =
		reflectionPrompts(
			d, report)
	report.NextActions =
		nextActions(
			d, report)
	report.SimilarTerms =
		similarityTerms(
			d)
	return report
}

func missingFields(
	d domain.Dream,
) []string {
	missing := []string{}
	if strings.TrimSpace(d.Content) == "" {
		missing = append(missing,
			"梦境正文")
	}
	if d.SleepHours <= 0 {
		missing = append(missing,
			"睡眠时长")
	}
	if d.Emotion == "" {
		missing = append(missing,
			"醒来情绪")
	}
	if len(d.Tags) == 0 {
		missing = append(missing,
			"梦中元素")
	}
	if len(d.Themes) == 0 {
		missing = append(missing,
			"主题证据")
	}
	return missing
}

func completeness(
	d domain.Dream,
	missing int,
) float64 {
	fields := 7.0
	if d.RememberDetail {
		fields++
	}
	return roundReflection((fields-float64(missing))/8, 3)
}

func reflectionEvidence(
	d domain.Dream,
) []ReflectionEvidence {
	result := []ReflectionEvidence{}
	for i := range d.Themes {
		theme := d.Themes[i]
		result = append(result,
			ReflectionEvidence{Kind: "主题", Name: theme.Name, Words: append([]string(nil), theme.Evidence...), Note: text.Hint(theme.Name)})
	}
	for i := range d.Tags {
		tag := d.Tags[i]
		result = append(result,
			ReflectionEvidence{Kind: "元素", Name: tag, Words: []string{tag}, Note: "可以回看它在梦中的位置、动作和情绪变化。"})
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Kind < result[j].Kind })
	return result
}

func reflectionPrompts(
	d domain.Dream,
	report ReflectionReport,
) []ReflectionPrompt {
	result := []ReflectionPrompt{}
	if report.Profile.Actions == 0 {
		result = append(result,
			ReflectionPrompt{Title: "补一处动作", Question: "梦里的人或你自己做了什么？动作发生前后有什么变化？", Reason: "当前叙事的动作信号较少。", Priority: "高"})
	}
	if report.Profile.Places == 0 {
		result = append(result,
			ReflectionPrompt{Title: "定位空间", Question: "你身处哪里？空间是熟悉的、封闭的，还是正在改变？", Reason: "地点线索有助于区分重复场景。", Priority: "中"})
	}
	if report.Profile.SensoryWords == 0 {
		result = append(result,
			ReflectionPrompt{Title: "补一个感官细节", Question: "醒来后最先留下的是颜色、声音、温度，还是身体感觉？", Reason: "感官词可以让下一次回看更容易恢复画面。", Priority: "中"})
	}
	if d.Emotion == domain.EmotionTense || d.Emotion == domain.EmotionAfraid {
		result = append(result,
			ReflectionPrompt{Title: "标记警觉来源", Question: "紧张或害怕是在梦中哪个瞬间升高的？醒来后还留在身体哪里？", Reason: "醒来情绪属于警觉类，需要保留发生节点。", Priority: "高"})
	}
	if len(d.Themes) > 0 {
		result = append(result,
			ReflectionPrompt{Title: "验证主题", Question: "主题证据词是否描述了你真正记得的画面，而不是事后联想？", Reason: "系统主题只是可复核线索，不是固定解释。", Priority: "高"})
	}
	if len(result) == 0 {
		result = append(result,
			ReflectionPrompt{Title: "保留原始版本", Question: "如果不解释这段梦，只保留一个最重要的画面，会是哪一个？", Reason: "完整记录适合先保存，再回看。", Priority: "低"})
	}
	return result
}

func nextActions(
	d domain.Dream,
	report ReflectionReport,
) []string {
	result := []string{}
	if len(report.Missing) > 0 {
		result = append(result,
			"下次醒来先补齐："+strings.Join(report.Missing, "、")+"。")
	}
	if len(d.Themes) > 0 {
		result = append(result,
			"一周后搜索“"+d.Themes[0].Name+"”，比较它和醒来情绪是否一起重复。")
	}
	if len(d.Tags) > 0 {
		result = append(result,
			"保留元素“"+d.Tags[0]+"”的原词，避免不同日期使用过多近义词。")
	}
	if len(result) == 0 {
		result = append(result,
			"继续记录三到五次，再比较主题、睡眠和清晰度。")
	}
	return result
}

func similarityTerms(
	d domain.Dream,
) []string {
	terms := []string{}
	for i := range d.Themes {
		theme := d.Themes[i]
		terms = append(terms,
			theme.Name)
	}
	for i := range d.Tags {
		tag := d.Tags[i]
		terms = append(terms, tag)
	}
	for word, count := range text.KeywordCounts(d.Title, d.Content) {
		if count >= 2 {
			terms = append(terms,
				word)
		}
	}
	sort.Strings(terms)
	return collections.Unique(terms)
}

func roundReflection(
	value float64,
	places int,
) float64 {
	power := 1.0
	for index := 0; index < places; index++ {
		power *= 10
	}
	return float64(int(value*power+0.5)) / power
}
