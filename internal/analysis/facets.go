package analysis

import (
	"dream117/internal/domain"
	"dream117/internal/text"
	"dream117/pkg/collections"
	"dream117/pkg/mathx"
)

type FacetReport struct {
	Emotions   []FacetValue
	Themes     []FacetValue
	Elements   []FacetValue
	Clarity    []FacetValue
	Sleep      []FacetValue
	Total      int
	WithDetail int
}

type FacetValue struct {
	Value string
	Label string
	Count int
	Share float64
	Hint  string
}

func BuildFacets(
	items []domain.Dream,
) FacetReport {
	report := FacetReport{Total: len(items)}
	emotions := EmotionCounts(items)
	themes := collections.CountNested(items, themeNames)
	elements := collections.CountNested(items, func(d domain.Dream) []string { return d.Tags })
	clarity := map[string]int{}
	sleep := map[string]int{}
	for i := range items {
		d := items[i]
		if d.RememberDetail {
			report.WithDetail++
		}
		clarity[clarityBand(d.Clarity)]++
		sleep[sleepBand(d.SleepHours)]++
	}
	report.Emotions =
		facetValues(
			emotions, len(items), "情绪")
	report.Themes = facetValues(
		themes, len(items), "主题")
	report.Elements =
		facetValues(
			elements, len(items), "元素")
	report.Clarity =
		orderedFacetValues(
			clarity, len(items), []string{"模糊 1-3", "中等 4-7", "清晰 8-10"})
	report.Sleep =
		orderedFacetValues(
			sleep, len(items), []string{"短睡眠 <6h", "适中 6-9h", "长睡眠 >=9h"})
	return report
}

func facetValues(
	counts map[string]int,
	total int,
	kind string,
) []FacetValue {
	ordered := collections.TopCounts(counts, 0)
	values := make(
		[]FacetValue,
		0,
		len(counts),
	)
	for i := range ordered {
		item := ordered[i]
		values = append(values,
			FacetValue{Value: item.Name, Label: item.Name, Count: item.Count, Share: mathx.Round(mathx.Ratio(item.Count, total), 3), Hint: kind + "筛选"})
	}
	return values
}

func orderedFacetValues(
	counts map[string]int,
	total int,
	order []string,
) []FacetValue {
	values := make(
		[]FacetValue,
		0,
		len(order),
	)
	for i := range order {
		value := order[i]
		values = append(values,
			FacetValue{Value: value, Label: value, Count: counts[value], Share: mathx.Round(mathx.Ratio(counts[value], total), 3), Hint: "结构化筛选"})
	}
	return values
}

func clarityBand(
	value int,
) string {
	switch {
	case value <= 3:
		return "模糊 1-3"
	case value <= 7:
		return "中等 4-7"
	default:
		return "清晰 8-10"
	}
}

func sleepBand(
	value float64,
) string {
	switch {
	case value < 6:
		return "短睡眠 <6h"
	case value < 9:
		return "适中 6-9h"
	default:
		return "长睡眠 >=9h"
	}
}

func FacetFilterValue(
	value string,
) (domain.DreamFilter, bool) {
	filter := domain.DreamFilter{}
	switch value {
	case "模糊 1-3":
		filter.MinimumClarity = 1
		filter.MaximumSleep = 0
	case "中等 4-7":
		filter.MinimumClarity = 4
	case "清晰 8-10":
		filter.MinimumClarity = 8
	case "短睡眠 <6h":
		filter.MaximumSleep = 6
	case "适中 6-9h", "长睡眠 >=9h":
		return filter, true
	default:
		return filter, false
	}
	return filter, true
}

func FacetExplanation(
	value FacetValue,
) string {
	if value.Count == 0 {
		return "当前窗口还没有这类记录。"
	}
	if value.Share >= 0.5 {
		return value.Label + "占当前窗口的一半以上。"
	}
	if value.Share >= 0.25 {
		return value.Label + "是一个值得持续观察的中等频率线索。"
	}
	return value.Label + "目前只在少量记录中出现。"
}

func FacetSummary(
	report FacetReport,
) []string {
	result := []string{}
	if report.Total == 0 {
		return []string{"先保存一条梦境，筛选项会随着记录自动出现。"}
	}
	if len(report.Emotions) >
		0 {
		result = append(result,
			"最常见情绪是“"+report.Emotions[0].Label+"”，共有 "+itoa(report.Emotions[0].Count)+" 条。")
	}
	if len(report.Themes) > 0 {
		result = append(result,
			"目前最常出现的主题线索是“"+report.Themes[0].Label+"”。")
	}
	result = append(result,
		"记得具体细节的记录有 "+itoa(report.WithDetail)+" 条。")
	return result
}

func RankedFacetTerms(
	items []domain.Dream,
) []FacetValue {
	counts := map[string]int{}
	for i := range items {
		d := items[i]
		collections.MergeCounts(counts, text.KeywordCounts(d.Title, d.Content))
	}
	return facetValues(counts, len(items), "叙事词")
}
