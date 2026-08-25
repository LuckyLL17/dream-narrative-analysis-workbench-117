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
	Band         string
	Label        string
	SharedTags   []string
	SharedThemes []string
	SharedWords  []string
	Reasons      []string
}

// SimilarDreamsResult bundles the ranked candidates, the shared band
// description and the effective limit so the HTTP layer and the overview
// cannot drift apart on boundary semantics or default paging.
type SimilarDreamsResult struct {
	Items []SimilarDream
	Bands []SimilarityBand
	Limit int
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
) SimilarDreamsResult {
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
		comparison.Band = SimilarityBandName(comparison.Score)
		comparison.Label = SimilarityLabel(comparison.Score)
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
	if limit <= 0 {
		limit = DefaultSimilarLimit
	}
	if len(result) > limit {
		result = result[:limit]
	}
	return SimilarDreamsResult{Items: result, Bands: SimilarityBands(), Limit: limit}
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
	Note  string
}

// DefaultSimilarLimit is the single source of truth for the default page
// size shared between the analysis layer, the service and the HTTP route so
// every entry point returns the same number of similar dreams.
const DefaultSimilarLimit = 8

// SimilarityBands describes the similarity segments using left-closed,
// right-open intervals: weak [0, 0.4), medium [0.4, 0.7), strong [0.7, 1].
// A score of exactly 0.4 lands in "部分相似" and a score of exactly 0.7
// lands in "强相似"; there is no gap between adjacent bands.
func SimilarityBands() []SimilarityBand {
	return []SimilarityBand{
		{Name: "weak", Min: 0, Max: 0.4, Label: "轻微重合", Note: "0 ≤ 分数 < 0.4：叙事重叠较少，仅作背景参考。"},
		{Name: "medium", Min: 0.4, Max: 0.7, Label: "部分相似", Note: "0.4 ≤ 分数 < 0.7：存在可追溯的共现线索，建议回看。"},
		{Name: "strong", Min: 0.7, Max: 1, Label: "强相似", Note: "0.7 ≤ 分数 ≤ 1：多条线索同时重合，优先对照比较。"},
	}
}

// SimilarityLabel returns the human-readable band label for a score using the
// same left-closed, right-open semantics as SimilarityBands. Scores at or
// above the strong band's minimum (including 1.0) are treated as "强相似";
// negative scores fall back to "轻微重合".
func SimilarityLabel(
	score float64,
) string {
	bands := SimilarityBands()
	for i := range bands {
		band := bands[i]
		if score >= band.Min && score < band.Max {
			return band.Label
		}
	}
	if len(bands) > 0 && score >= bands[len(bands)-1].Min {
		return bands[len(bands)-1].Label
	}
	return bands[0].Label
}

// SimilarityBandName returns the machine band identifier for a score, mirroring
// SimilarityLabel so callers can label and classify from one source.
func SimilarityBandName(
	score float64,
) string {
	bands := SimilarityBands()
	for i := range bands {
		band := bands[i]
		if score >= band.Min && score < band.Max {
			return band.Name
		}
	}
	if len(bands) > 0 && score >= bands[len(bands)-1].Min {
		return bands[len(bands)-1].Name
	}
	return bands[0].Name
}
