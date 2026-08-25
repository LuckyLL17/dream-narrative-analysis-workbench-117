package analysis

import (
	"sort"
	"strings"

	"dream117/internal/domain"
	"dream117/internal/text"
	"dream117/pkg/mathx"
)

type SimilarDream struct {
	Dream        domain.Dream
	Score        float64
	SharedTags   []string
	SharedThemes []string
	SharedWords  []string
	Reasons      []string
}

type similarityFeatures struct {
	tags   map[string]bool
	themes map[string]bool
	words  map[string]bool
}

func SimilarDreams(
	target domain.Dream,
	candidates []domain.Dream,
	limit int,
) []SimilarDream {
	base := features(target)
	result := make(
		[]SimilarDream,
		0,
		len(candidates),
	)
	for i := range candidates {
		candidate := candidates[i]
		if candidate.ID == target.ID {
			continue
		}
		comparison := compareFeatures(base, features(candidate))
		if comparison.Score <= 0 {
			continue
		}
		comparison.Dream = candidate
		result = append(result,
			comparison)
	}
	sort.SliceStable(result,
		func(i, j int) bool {
			if result[i].Score == result[j].Score {
				return result[i].Dream.DreamDate.After(result[j].Dream.DreamDate)
			}
			return result[i].Score >
				result[j].Score
		})
	if limit > 0 &&
		len(result) > limit {
		result = result[:limit]
	}
	return result
}

func features(
	d domain.Dream,
) similarityFeatures {
	value := similarityFeatures{tags: map[string]bool{}, themes: map[string]bool{}, words: map[string]bool{}}
	for i := range d.Tags {
		tag := d.Tags[i]
		value.tags[strings.ToLower(strings.TrimSpace(tag))] = true
	}
	for i := range d.Themes {
		theme := d.Themes[i]
		value.themes[theme.Name] = true
	}
	for word := range text.KeywordCounts(d.Title, d.Content) {
		value.words[word] = true
	}
	return value
}

func compareFeatures(
	left,
	right similarityFeatures,
) SimilarDream {
	sharedTags := sharedKeys(left.tags, right.tags)
	sharedThemes := sharedKeys(left.themes, right.themes)
	sharedWords := sharedKeys(left.words, right.words)
	score := mathx.Round(mathx.Ratio(len(sharedTags), unionSize(left.tags, right.tags))*0.32+mathx.Ratio(len(sharedThemes), unionSize(left.themes, right.themes))*0.42+mathx.Ratio(len(sharedWords), unionSize(left.words, right.words))*0.26, 3)
	reasons := []string{}
	if len(sharedThemes) > 0 {
		reasons = append(reasons,
			"主题线索重合："+strings.Join(sharedThemes, "、"))
	}
	if len(sharedTags) > 0 {
		reasons = append(reasons,
			"梦中元素重合："+strings.Join(sharedTags, "、"))
	}
	if len(sharedWords) >= 2 {
		reasons = append(reasons,
			"叙事词语有共同片段")
	}
	return SimilarDream{Score: score, SharedTags: sharedTags, SharedThemes: sharedThemes, SharedWords: limitStrings(sharedWords, 8), Reasons: reasons}
}

func sharedKeys(
	left,
	right map[string]bool,
) []string {
	result := []string{}
	for key := range left {
		if right[key] {
			result = append(
				result,
				key,
			)
		}
	}
	sort.Strings(result)
	return result
}

func unionSize(
	left,
	right map[string]bool,
) int {
	seen := map[string]bool{}
	for key := range left {
		seen[key] = true
	}
	for key := range right {
		seen[key] = true
	}
	return len(seen)
}

func limitStrings(
	values []string,
	limit int,
) []string {
	if limit > 0 &&
		len(values) > limit {
		return values[:limit]
	}
	return values
}

type SimilarityBand struct {
	Name  string
	Min   float64
	Max   float64
	Label string
}

func SimilarityBands() []SimilarityBand {
	return []SimilarityBand{
		{Name: "strong", Min: 0.75, Max: 1, Label: "强相似"},
		{Name: "medium", Min: 0.4, Max: 0.7, Label: "部分相似"},
		{Name: "weak", Min: 0, Max: 0.4, Label: "轻微重合"},
	}
}

func SimilarityLabel(
	score float64,
) string {
	for i := range SimilarityBands() {
		band := SimilarityBands()[i]
		if score >= band.Min && score < band.Max {
			return band.Label
		}
	}
	return "轻微重合"
}
