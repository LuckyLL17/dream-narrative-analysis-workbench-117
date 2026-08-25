package text

import (
	"dream117/internal/domain"
	"dream117/pkg/collections"
)

type Keyword struct {
	Word  string
	Count int
	Score float64
}

func Keywords(
	title,
	content string,
) []Keyword {
	counts := KeywordCounts(title, content)
	items := make(
		[]Keyword,
		0,
		len(counts),
	)
	for word := range counts {
		count := counts[word]
		items = append(items,
			Keyword{Word: word, Count: count, Score: float64(count) + titleKeywordBonus(word, title)})
	}
	return collections.SortByScoreName(items,
		func(item Keyword) float64 { return item.Score },
		func(item Keyword) string { return item.Word })
}

func titleKeywordBonus(word, title string) float64 {
	if title == "" {
		return 0
	}
	if KeywordCounts(title, "")[word] > 0 {
		return 2.5
	}
	return 0
}

func KeywordCounts(
	title,
	content string,
) map[string]int {
	counts := collections.Counter()
	tokens := Tokens(title, content)
	for i := range tokens {
		token := tokens[i]
		counts[token]++
	}
	return counts
}

func EmotionIntensity(
	emotion domain.Emotion,
	clarity int,
) float64 {
	base := map[domain.Emotion]float64{domain.EmotionHappy: 0.6, domain.EmotionTense: 0.85, domain.EmotionConfused: 0.55, domain.EmotionCalm: 0.3, domain.EmotionAfraid: 0.95, domain.EmotionSurprised: 0.75}[emotion]
	if clarity < 1 {
		clarity = 1
	}
	if clarity > 10 {
		clarity = 10
	}
	return base * (0.55 + float64(clarity)/20)
}
