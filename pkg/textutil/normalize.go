package textutil

import (
	"strings"
	"unicode"
)

func Clean(
	value string,
) string {
	var builder strings.Builder
	lastSpace := false
	for _, runeValue := range strings.TrimSpace(value) {
		if unicode.IsSpace(runeValue) || unicode.IsPunct(runeValue) {
			if !lastSpace {
				builder.WriteRune(' ')
			}
			lastSpace = true
			continue
		}
		builder.WriteRune(unicode.ToLower(runeValue))
		lastSpace = false
	}
	return strings.TrimSpace(builder.String())
}

func ContainsAny(
	value string,
	terms []string,
) []string {
	cleaned := Clean(value)
	matched := make(
		[]string,
		0,
		len(terms),
	)
	for i := range terms {
		term := terms[i]
		if term != "" && strings.Contains(cleaned, Clean(term)) {
			matched = append(matched,
				term)
		}
	}
	return matched
}

func Unique(
	values []string,
) []string {
	seen := make(
		map[string]struct{}, len(values))
	result := make(
		[]string,
		0,
		len(values),
	)
	for i := range values {
		value := values[i]
		value = strings.TrimSpace(
			value)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(
			result,
			value,
		)
	}
	return result
}
