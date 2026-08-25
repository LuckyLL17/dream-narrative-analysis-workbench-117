package analysis

import (
	"dream117/internal/domain"
	"dream117/internal/text"
)

func WordCloud(
	items []domain.Dream,
	limit int,
) []domain.WordStat {
	// RankedWords already returns words ordered by the unified weight; do not
	// re-sort by raw count here, otherwise the title semantic bonus computed
	// upstream is discarded and the wordcloud diverges from the report and
	// sentence summary.
	return RankedWords(items, limit)
}

// WordCloudSentence renders the word list as a frequency-weighted, weight-
// ordered summary sentence. It consumes the same ordering produced by
// RankedWords / WordCloud, so the sentence and the wordcloud always agree.
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
