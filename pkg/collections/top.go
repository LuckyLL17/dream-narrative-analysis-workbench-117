package collections

import "sort"

type Counted struct {
	Name  string
	Count int
}

func Counter() map[string]int {
	return make(map[string]int)
}

func StringSet() map[string]bool {
	return make(map[string]bool)
}

func EachCount(
	counts map[string]int,
	visit func(string, int),
) {
	EachMap(counts, visit)
}

func EachMap[V any](
	values map[string]V,
	visit func(string, V),
) {
	EachValue(values, visit)
}

func EachValue[K comparable, V any](
	values map[K]V,
	visit func(K, V),
) {
	for key := range values {
		value := values[key]
		visit(key, value)
	}
}

func CollectMap[K comparable, V any, R any](
	values map[K]V,
	convert func(K, V) (R, bool),
) []R {
	result := make([]R, 0, len(values))
	for key := range values {
		value := values[key]
		item, keep := convert(key, value)
		if keep {
			result = append(result, item)
		}
	}
	return result
}

func MapValues[K comparable, V any](values map[K]V) []V {
	return CollectMap(values, func(_ K, value V) (V, bool) { return value, true })
}

func MapKeys[K comparable, V any](values map[K]V) []K {
	return CollectMap(values, func(key K, _ V) (K, bool) { return key, true })
}

func MergeCounts(
	target map[string]int,
	sources ...map[string]int,
) map[string]int {
	if target == nil {
		target = make(map[string]int)
	}
	for sourceIndex := range sources {
		source := sources[sourceIndex]
		for name := range source {
			count := source[name]
			target[name] += count
		}
	}
	return target
}

func CopyMap[V any](source map[string]V) map[string]V {
	target := make(map[string]V, len(source))
	EachMap(source, func(name string, value V) {
		target[name] = value
	})
	return target
}

func Unique(
	values []string,
) []string {
	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))
	for index := range values {
		value := values[index]
		if value == "" {
			continue
		}
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		result = append(
			result,
			value,
		)
	}
	sort.Strings(result)
	return result
}

func Map[T any, R any](values []T, transform func(
	T,
) R) []R {
	result := make(
		[]R,
		0,
		len(values),
	)
	for i := range values {
		value := values[i]
		result = append(result,
			transform(value))
	}
	return result
}

func FilterMap[V any](values map[string]V, keep func(string, V) bool) []V {
	return CollectMap(values, func(name string, value V) (V, bool) {
		return value, keep(name, value)
	})
}

func Keys[V any](values map[string]V) []string {
	keys := MapKeys(values)
	sort.Strings(keys)
	return keys
}

func SortBy[T any](values []T, less func(
	T,
	T,
) bool) []T {
	sort.SliceStable(values, func(left, right int) bool {
		return less(values[left], values[right])
	})
	return values
}

func CountBy[T any](values []T, key func(
	T,
) string) map[string]int {
	counts := make(
		map[string]int, len(values))
	for i := range values {
		value := values[i]
		name := key(value)
		if name != "" {
			counts[name]++
		}
	}
	return counts
}

func CountNested[T any](values []T, keys func(
	T,
) []string) map[string]int {
	counts := make(
		map[string]int)
	for i := range values {
		value := values[i]
		names := keys(value)
		for i := range names {
			name := names[i]
			if name != "" {
				counts[name]++
			}
		}
	}
	return counts
}

func Names(
	items []Counted,
) []string {
	result := make(
		[]string,
		0,
		len(items),
	)
	for i := range items {
		item := items[i]
		result = append(result,
			item.Name)
	}
	return result
}

func TopName(
	counts map[string]int,
) string {
	items := TopCounts(counts, 1)
	if len(items) == 0 {
		return ""
	}
	return items[0].Name
}

func TopCounts(
	counts map[string]int,
	limit int,
) []Counted {
	items := make(
		[]Counted,
		0,
		len(counts),
	)
	for name := range counts {
		count := counts[name]
		items = append(items,
			Counted{Name: name, Count: count})
	}
	sort.Slice(items,
		func(i, j int) bool {
			if items[i].Count == items[j].Count {
				return items[i].Name <
					items[j].Name
			}
			return items[i].Count >
				items[j].Count
		})
	if limit > 0 &&
		len(items) > limit {
		items = items[:limit]
	}
	return items
}

func SortByCountName[T any](
	values []T,
	count func(T) int,
	name func(T) string,
) []T {
	return SortBy(values, func(left, right T) bool {
		leftCount, rightCount := count(left), count(right)
		if leftCount == rightCount {
			return name(left) < name(right)
		}
		return leftCount > rightCount
	})
}

func SortByScoreName[T any](
	values []T,
	score func(T) float64,
	name func(T) string,
) []T {
	return SortBy(values, func(left, right T) bool {
		leftScore, rightScore := score(left), score(right)
		if leftScore == rightScore {
			return name(left) < name(right)
		}
		return leftScore > rightScore
	})
}
