package analysis

import (
	"dream117/internal/domain"
	"dream117/internal/text"
	"dream117/pkg/collections"
	"dream117/pkg/mathx"
)

type SleepGroup struct {
	Name           string
	Minimum        float64
	Maximum        float64
	Count          int
	AverageSleep   float64
	AverageClarity float64
	AverageMood    float64
	TopEmotion     string
	Themes         []string
}

func SleepContrast(
	items []domain.Dream,
) []SleepGroup {
	groups := []SleepGroup{{Name: "短睡眠", Maximum: 6}, {Name: "适中", Minimum: 6, Maximum: 9}, {Name: "长睡眠", Minimum: 9}}
	for index := range groups {
		clarity, sleep, mood := []float64{}, []float64{}, []float64{}
		emotions := map[string]int{}
		themes := map[string]int{}
		for i := range items {
			d := items[i]
			if !inGroup(d.SleepHours, groups[index]) {
				continue
			}
			groups[index].Count++
			clarity = append(clarity,
				float64(d.Clarity))
			sleep = append(sleep,
				d.SleepHours)
			mood = append(mood,
				text.EmotionIntensity(d.Emotion, d.Clarity))
			emotions[string(d.Emotion)]++
			for i := range d.Themes {
				theme := d.Themes[i]
				themes[theme.Name]++
			}
		}
		groups[index].AverageSleep = mathx.Round(mathx.Average(sleep), 2)
		groups[index].AverageClarity = mathx.Round(mathx.Average(clarity), 2)
		groups[index].AverageMood = mathx.Round(mathx.Average(mood), 2)
		groups[index].TopEmotion = collections.TopName(emotions)
		groups[index].Themes = topNames(themes, 4)
	}
	return groups
}

func inGroup(
	value float64,
	group SleepGroup,
) bool {
	if value < group.Minimum {
		return false
	}
	return group.Maximum == 0 || value <
		group.Maximum
}
func topNames(
	counts map[string]int,
	limit int,
) []string {
	return collections.Names(collections.TopCounts(counts, limit))
}

func SleepAdvice(
	groups []SleepGroup,
) []string {
	result := []string{}
	if len(groups) < 3 {
		return result
	}
	if groups[0].Count > 0 && groups[2].Count > 0 && groups[0].AverageClarity+1 < groups[2].AverageClarity {
		result = append(result,
			"当前记录显示，长睡眠组的平均清晰度更高；可以继续观察是否稳定出现。")
	}
	if groups[0].AverageMood > groups[1].AverageMood && groups[0].Count > 1 {
		result = append(result,
			"短睡眠记录的情绪强度偏高，建议在记录中补充睡前环境和醒来时的身体感受。")
	}
	if len(result) == 0 {
		result = append(result,
			"样本还不足以形成稳定的睡眠建议，继续记录比提前解释更重要。")
	}
	return result
}
