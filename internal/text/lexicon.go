package text

import (
	"strings"

	"dream117/internal/domain"
	"dream117/pkg/collections"
)

type LexiconEntry struct {
	Term        string
	Kind        string
	Themes      []string
	Emotions    []domain.Emotion
	Description string
	Examples    []string
}

func BuildLexicon() []LexiconEntry {
	entries := make(
		[]LexiconEntry, 0, len(elementRules)+len(themeRules))
	eachElementTerm(func(rule ElementRule, term string) {
		entries = append(entries,
			LexiconEntry{Term: term, Kind: string(rule.Kind), Description: elementDescription(rule.Kind), Examples: []string{"人物、地点和动作可以帮助回看梦境的叙事结构"}})
	})
	for i := range themeRules {
		rule := themeRules[i]
		entries = append(entries,
			LexiconEntry{Term: rule.Name, Kind: "主题线索", Themes: []string{rule.Name}, Description: rule.Hint, Examples: append([]string(nil), rule.Evidence...)})
	}
	collections.SortBy(
		entries,
		func(left, right LexiconEntry) bool {
			return left.Term < right.Term
		},
	)
	return entries
}

func FindLexicon(
	value string,
) []LexiconEntry {
	query := strings.ToLower(strings.TrimSpace(value))
	result := []LexiconEntry{}
	for _, entry := range BuildLexicon() {
		if query == "" || strings.Contains(strings.ToLower(entry.Term+" "+entry.Description), query) {
			result = append(
				result,
				entry,
			)
		}
	}
	return result
}
func elementDescription(
	kind domain.ElementKind,
) string {
	switch kind {
	case domain.ElementPerson:
		return "记录人物关系和互动位置"
	case domain.ElementAnimal:
		return "记录动物形象和行动方式"
	case domain.ElementPlace:
		return "记录空间、边界和方向感"
	case domain.ElementObject:
		return "记录承载记忆的具体物件"
	case domain.ElementAction:
		return "记录梦中的身体动作和事件"
	case domain.ElementColor:
		return "记录视觉色彩和氛围"
	default:
		return "记录天气、自然和环境变化"
	}
}
