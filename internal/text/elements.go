package text

import (
	"strings"

	"dream117/internal/domain"
	"dream117/pkg/textutil"
)

type ElementRule struct {
	Kind  domain.ElementKind
	Terms []string
}

var elementRules = []ElementRule{
	{Kind: domain.ElementPerson, Terms: []string{"朋友", "母亲", "父亲", "老师", "同学", "陌生人", "孩子", "老人", "医生", "家人"}},
	{Kind: domain.ElementAnimal, Terms: []string{"猫", "狗", "鸟", "蛇", "马", "鱼", "蝴蝶", "狼", "熊", "兔子"}},
	{Kind: domain.ElementPlace, Terms: []string{"学校", "家里", "车站", "医院", "森林", "海边", "楼顶", "房间", "街道", "电梯", "考场"}},
	{Kind: domain.ElementObject, Terms: []string{"手机", "钥匙", "镜子", "书", "门", "车", "雨伞", "箱子", "钟", "牙齿"}},
	{Kind: domain.ElementAction, Terms: []string{"奔跑", "追赶", "坠落", "飞行", "考试", "迟到", "寻找", "躲藏", "游泳", "说话"}},
	{Kind: domain.ElementColor, Terms: []string{"红色", "蓝色", "黑色", "白色", "金色", "绿色", "紫色", "灰色"}},
	{Kind: domain.ElementNature, Terms: []string{"雨", "雪", "风", "海", "河", "月亮", "太阳", "火", "雾", "雷"}},
}

func eachElementTerm(visit func(ElementRule, string)) {
	for ruleIndex := range elementRules {
		rule := elementRules[ruleIndex]
		for termIndex := range rule.Terms {
			term := rule.Terms[termIndex]
			visit(rule, term)
		}
	}
}

func RecommendElements(
	title,
	content string,
) []domain.Element {
	value := strings.ToLower(textutil.Clean(title + " " + content))
	result := make(
		[]domain.Element, 0)
	eachElementTerm(func(rule ElementRule, term string) {
		if strings.Contains(value, strings.ToLower(term)) {
			result = append(result,
				domain.Element{Name: term, Kind: rule.Kind})
		}
	})
	return result
}

func Rules() []ElementRule { return append([]ElementRule(nil), elementRules...) }

func KindFor(
	name string,
) domain.ElementKind {
	kind := domain.ElementObject
	eachElementTerm(func(rule ElementRule, term string) {
		if term == name {
			kind = rule.Kind
		}
	})
	return kind
}
