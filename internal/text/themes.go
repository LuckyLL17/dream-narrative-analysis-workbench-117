package text

import (
	"strings"

	"dream117/internal/domain"
	"dream117/pkg/textutil"
)

type ThemeRule struct {
	Name     string
	Evidence []string
	Weight   float64
	Hint     string
}

var themeRules = []ThemeRule{
	{Name: "被追赶", Evidence: []string{"追赶", "逃跑", "奔跑", "躲藏", "追我"}, Weight: 1.0, Hint: "身体处于持续警觉，梦境里出现了需要摆脱的压力线索。"},
	{Name: "坠落", Evidence: []string{"坠落", "掉下", "高处", "楼顶", "悬崖", "失重"}, Weight: 1.1, Hint: "控制感或稳定感可能是这段叙事的核心。"},
	{Name: "飞行", Evidence: []string{"飞行", "飞起来", "天空", "翅膀", "漂浮"}, Weight: 0.9, Hint: "叙事包含空间突破与自由感线索。"},
	{Name: "考试与准备", Evidence: []string{"考试", "考场", "试卷", "作业", "答题", "迟到"}, Weight: 1.0, Hint: "评价、准备程度或错过节点的担忧值得关注。"},
	{Name: "迷路", Evidence: []string{"迷路", "找不到", "方向", "地图", "陌生街道", "走散"}, Weight: 0.95, Hint: "梦境反复围绕方向感、选择或关系距离展开。"},
	{Name: "与逝者对话", Evidence: []string{"逝者", "去世", "葬礼", "墓地", "已故", "回来"}, Weight: 1.2, Hint: "记忆、告别或未说完的话进入了叙事。"},
	{Name: "牙齿掉落", Evidence: []string{"牙齿", "掉牙", "牙掉", "口腔"}, Weight: 1.15, Hint: "身体感受和表达能力是这条线索的重要入口。"},
	{Name: "水域与淹没", Evidence: []string{"海", "河", "游泳", "淹没", "水下", "洪水"}, Weight: 0.9, Hint: "情绪强度通过水域空间被放大，适合结合醒来情绪观察。"},
}

func DetectThemes(
	title,
	content string,
) []domain.ThemeHit {
	value := strings.ToLower(textutil.Clean(title + " " + content))
	hits := make(
		[]domain.ThemeHit, 0)
	for i := range themeRules {
		rule := themeRules[i]
		evidence := make([]string, 0)
		for i := range rule.Evidence {
			term := rule.Evidence[i]
			if strings.Contains(value, strings.ToLower(term)) {
				evidence = append(evidence,
					term)
			}
		}
		if len(evidence) == 0 {
			continue
		}
		confidence := float64(len(evidence)) / float64(len(rule.Evidence)) * rule.Weight
		if confidence > 1 {
			confidence = 1
		}
		hits = append(hits,
			domain.ThemeHit{Name: rule.Name, Confidence: confidence, Evidence: evidence})
	}
	return hits
}

func ThemeRules() []ThemeRule { return append([]ThemeRule(nil), themeRules...) }

func Hint(
	name string,
) string {
	for i := range themeRules {
		rule := themeRules[i]
		if rule.Name == name {
			return rule.Hint
		}
	}
	return ""
}
