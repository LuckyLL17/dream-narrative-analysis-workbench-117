package text

import (
	"strings"

	"dream117/internal/domain"
)

func SearchScore(
	d domain.Dream,
	query string,
) float64 {
	query = strings.ToLower(
		strings.TrimSpace(query))
	if query == "" {
		return 0
	}
	value := strings.ToLower(d.Title + " " + d.Content + " " + strings.Join(d.Tags, " "))
	score := 0.0
	if strings.Contains(strings.ToLower(d.Title), query) {
		score += 3
	}
	if strings.Contains(value, query) {
		score += 1
	}
	for i := range d.Themes {
		theme := d.Themes[i]
		if strings.Contains(strings.ToLower(theme.Name), query) {
			score += 2
		}
	}
	return score
}

func MoodBucket(
	emotion domain.Emotion,
) string {
	switch emotion {
	case domain.EmotionHappy, domain.EmotionCalm:
		return "舒缓"
	case domain.EmotionTense, domain.EmotionAfraid:
		return "警觉"
	default:
		return "探索"
	}
}
