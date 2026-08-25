package text

import (
	"strings"
)

type NarrativeProfile struct {
	Characters   int
	Sentences    int
	Questions    int
	Actions      int
	Places       int
	SensoryWords int
	Density      float64
	Opening      string
	Closing      string
	Signals      []string
}

func Profile(
	title,
	content string,
) NarrativeProfile {
	value := strings.TrimSpace(title + "。" + content)
	profile := NarrativeProfile{Characters: len([]rune(value)), Sentences: sentenceCount(value), Questions: strings.Count(value, "？") + strings.Count(value, "?")}
	for i := range Tokens(title, content) {
		token :=
			Tokens(title, content)[i]
		if containsAction(token) {
			profile.Actions++
		}
		if containsPlace(token) {
			profile.Places++
		}
		if containsSensory(token) {
			profile.SensoryWords++
		}
	}
	if profile.Sentences > 0 {
		profile.Density = float64(
			len(Tokens(title, content))) / float64(profile.Sentences)
	}
	profile.Opening, profile.Closing = excerpt(value, true), excerpt(value, false)
	profile.Signals =
		narrativeSignals(
			profile)
	return profile
}

func sentenceCount(
	value string,
) int {
	count := 0
	for _, r := range value {
		if strings.ContainsRune("。！？!?；;", r) {
			count++
		}
	}
	if count == 0 && strings.TrimSpace(value) != "" {
		return 1
	}
	return count
}
func containsAction(
	value string,
) bool {
	return strings.ContainsAny(value, "跑走飞落追找躲说游考")
}
func containsPlace(
	value string,
) bool {
	found := false
	eachElementTerm(func(rule ElementRule, term string) {
		if rule.Kind == "地点" && strings.Contains(value, term) {
			found = true
		}
	})
	return found
}
func containsSensory(
	value string,
) bool {
	return strings.ContainsAny(value, "明暗冷热响静颜色光")
}
func excerpt(
	value string,
	opening bool,
) string {
	runes := []rune(strings.TrimSpace(value))
	if len(runes) <= 28 {
		return string(runes)
	}
	if opening {
		return string(runes[:28]) + "…"
	}
	return "…" + string(runes[len(runes)-28:])
}
func narrativeSignals(
	profile NarrativeProfile,
) []string {
	result := []string{}
	if profile.Actions >= 3 {
		result = append(result,
			"动作密度较高")
	}
	if profile.Places >= 2 {
		result = append(result,
			"空间切换明显")
	}
	if profile.SensoryWords >= 2 {
		result = append(result,
			"感官画面丰富")
	}
	if profile.Questions > 0 {
		result = append(result,
			"叙事包含疑问")
	}
	if len(result) == 0 {
		result = append(result,
			"片段式记录")
	}
	return result
}
