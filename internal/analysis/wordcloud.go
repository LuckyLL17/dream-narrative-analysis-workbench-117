package analysis

import (
	"dream117/internal/domain"
	"dream117/internal/text"
	"dream117/pkg/collections"
)

func WordCloud(
	items []domain.Dream,
	limit int,
) []domain.WordStat {
	words := RankedWords(items, limit)
	collections.SortBy(
		words,
		func(left, right domain.WordStat) bool {
			return left.Weight > right.Weight
		},
	)
	return words
}

func WordCloudSentence(
	words []domain.WordStat,
) string {
	result := ""
	for i := range words {
		word := words[i]
		for index := 0; index < word.Count; index++ {
			if result != "" {
				result += "、"
			}
			result += word.Word
		}
	}
	return result
}

func ExtractWordSet(
	items []domain.Dream,
) map[string]struct{} {
	result := map[string]struct{}{}
	for i := range items {
		d := items[i]
		for word := range text.KeywordCounts(d.Title, d.Content) {
			result[word] = struct{}{}
		}
	}
	return result
}
