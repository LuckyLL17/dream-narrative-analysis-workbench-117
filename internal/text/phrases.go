package text

import (
	"sort"
	"strings"
	"unicode"

	"dream117/internal/domain"
	"dream117/pkg/collections"
	"dream117/pkg/textutil"
)

type Phrase struct {
	Text         string
	Count        int
	DistinctDays int
	Weight       float64
	Category     string
}

func Phrases(
	title,
	content string,
	limit int,
) []Phrase {
	result := rankPhrases(phraseCounts([]phraseSource{{title: title, content: content, day: "single"}}))
	return limitPhrases(result, limit)
}

func PhrasesAcrossDreams(
	items []domain.Dream,
	limit int,
) []Phrase {
	sources := make(
		[]phraseSource,
		0,
		len(items),
	)
	for i := range items {
		d := items[i]
		sources = append(sources,
			phraseSource{title: d.Title, content: d.Content, day: d.DreamDate.Format("2006-01-02")})
	}
	return limitPhrases(rankPhrases(phraseCounts(sources)), limit)
}

type phraseSource struct {
	title   string
	content string
	day     string
}

type phraseAggregate struct {
	count int
	days  map[string]bool
}

func phraseCounts(
	sources []phraseSource,
) map[string]phraseAggregate {
	counts := make(map[string]phraseAggregate)
	for i := range sources {
		source := sources[i]
		clean := textutil.Clean(source.title + " " + source.content)
		runes := []rune(clean)
		for index := 0; index < len(runes); index++ {
			if !unicode.IsLetter(runes[index]) && !unicode.IsDigit(runes[index]) {
				continue
			}
			for width := 2; width <= 4 && index+width <= len(runes); width++ {
				part := strings.TrimSpace(string(runes[index : index+width]))
				if !validPhrase(part) {
					continue
				}
				current := counts[part]
				current.count++
				if current.days == nil {
					current.days = collections.StringSet()
				}
				current.days[source.day] = true
				counts[part] = current
			}
		}
	}
	return counts
}

func rankPhrases(
	counts map[string]phraseAggregate,
) []Phrase {
	result := make(
		[]Phrase,
		0,
		len(counts),
	)
	collections.EachMap(counts, func(word string, aggregate phraseAggregate) {
		result = append(result,
			Phrase{Text: word, Count: aggregate.count, DistinctDays: len(aggregate.days), Weight: float64(aggregate.count) * phraseWeight(word) * (1 + float64(len(aggregate.days))/10), Category: phraseCategory(word)})
	})
	sort.Slice(result,
		func(i, j int) bool {
			if result[i].Weight == result[j].Weight {
				return result[i].Text <
					result[j].Text
			}
			return result[i].Weight >
				result[j].Weight
		})
	return result
}

func limitPhrases(
	result []Phrase,
	limit int,
) []Phrase {
	if limit > 0 &&
		len(result) > limit {
		return result[:limit]
	}
	return result
}

func validPhrase(
	value string,
) bool {
	if len([]rune(value)) < 2 {
		return false
	}
	if _, found := stopWords[value]; found {
		return false
	}
	return strings.TrimSpace(value) != ""
}
func phraseWeight(
	value string,
) float64 {
	weight := 1.0
	for i := range themeRules {
		rule := themeRules[i]
		for i := range rule.Evidence {
			term := rule.Evidence[i]
			if strings.Contains(value, term) {
				weight += rule.Weight * 0.18
			}
		}
	}
	return weight
}
func phraseCategory(
	value string,
) string {
	if strings.Contains(value, "追") || strings.Contains(value, "逃") || strings.Contains(value, "跑") {
		return "动作"
	}
	if strings.Contains(value, "海") || strings.Contains(value, "雨") || strings.Contains(value, "水") {
		return "环境"
	}
	return "叙事"
}

func NormalizeSearchQuery(
	value string,
) []string {
	value = textutil.Clean(value)
	if value == "" {
		return nil
	}
	return strings.FieldsFunc(value, func(r rune) bool { return unicode.IsSpace(r) || strings.ContainsRune(",，。！？、/", r) })
}

func QueryMatch(
	title,
	content,
	query string,
) (bool, []string) {
	terms := NormalizeSearchQuery(query)
	if len(terms) == 0 {
		return true, nil
	}
	value := strings.ToLower(title + " " + content)
	matched := []string{}
	for i := range terms {
		term := terms[i]
		if strings.Contains(value, strings.ToLower(term)) {
			matched = append(matched,
				term)
		}
	}
	return len(matched) == len(terms), matched
}
