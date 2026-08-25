package analysis

import (
	"sort"

	"dream117/internal/domain"
	"dream117/pkg/collections"
	"dream117/pkg/mathx"
)

type ThemePair struct {
	Left       string
	Right      string
	Count      int
	Jaccard    float64
	SharedMood string
}

type ThemeTimelinePoint struct {
	Period string
	Counts map[string]int
}

func ThemeCooccurrence(
	items []domain.Dream,
	limit int,
) []ThemePair {
	counts := collections.Counter()
	pairs := collections.Counter()
	moods := map[string]map[string]int{}
	for i := range items {
		d := items[i]
		names := uniqueThemeNames(d.Themes)
		for i := range names {
			name := names[i]
			counts[name]++
			if moods[name] == nil {
				moods[name] = map[string]int{}
			}
			moods[name][string(d.Emotion)]++
		}
		for left := 0; left < len(names); left++ {
			for right := left + 1; right < len(names); right++ {
				// Two dreams describing the same set of themes in a different
				// order are the same unordered pair, so the key must not depend
				// on the slice order. Canonicalise the two names before hashing.
				key := pairKey(names[left], names[right])
				pairs[key]++
			}
		}
	}
	result := make(
		[]ThemePair,
		0,
		len(pairs),
	)
	for k, n := range pairs {
		parts := splitPair(k)
		union := counts[parts[0]] + counts[parts[1]] - n
		result = append(result,
			ThemePair{Left: parts[0], Right: parts[1], Count: n, Jaccard: mathx.Round(mathx.Ratio(n, union), 3), SharedMood: collections.TopName(collections.MergeCounts(nil, moods[parts[0]], moods[parts[1]]))})
	}
	sort.Slice(result,
		func(i, j int) bool {
			if result[i].Count == result[j].Count {
				return result[i].Jaccard >
					result[j].Jaccard
			}
			return result[i].Count >
				result[j].Count
		})
	if limit > 0 &&
		len(result) > limit {
		result = result[:limit]
	}
	return result
}

func ThemeTimeline(
	items []domain.Dream,
	months int,
) []ThemeTimelinePoint {
	groups := map[string]map[string]int{}
	for i := range items {
		d := items[i]
		period := d.DreamDate.Format("2006-01")
		if groups[period] == nil {
			groups[period] = map[string]int{}
		}
		for i := range d.Themes {
			theme := d.Themes[i]
			groups[period][theme.Name]++
		}
	}
	periods := make(
		[]string,
		0,
		len(groups),
	)
	for period := range groups {
		periods = append(periods,
			period)
	}
	sort.Strings(periods)
	if months > 0 &&
		len(periods) > months {
		periods = periods[len(periods)-months:]
	}
	result := make(
		[]ThemeTimelinePoint,
		0,
		len(periods),
	)
	for i := range periods {
		period := periods[i]
		result = append(result,
			ThemeTimelinePoint{Period: period, Counts: groups[period]})
	}
	return result
}

func uniqueThemeNames(
	hits []domain.ThemeHit,
) []string {
	result := make([]string, 0, len(hits))
	seen := map[string]bool{}
	for i := range hits {
		name := hits[i].Name
		if name == "" || seen[name] {
			continue
		}
		seen[name] = true
		result = append(result, name)
	}
	return result
}
func splitPair(
	value string,
) []string {
	for index := range value {
		if value[index] == '\x00' {
			return []string{value[:index], value[index+1:]}
		}
	}
	return []string{value, ""}
}

// pairKey builds a canonical, order-independent key for an unordered pair of
// theme names. Whichever name comes first in the dream's slice, the same pair
// always maps to the same key, so "被追赶—迷路" and "迷路—被追赶" merge into a
// single co-occurrence edge instead of producing two reversed Count=1 edges.
func pairKey(
	left,
	right string,
) string {
	if left <= right {
		return left + "\x00" + right
	}
	return right + "\x00" + left
}
