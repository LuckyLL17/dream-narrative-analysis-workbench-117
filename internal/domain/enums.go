package domain

import "fmt"

type Emotion string

const (
	EmotionHappy     Emotion = "开心"
	EmotionTense     Emotion = "紧张"
	EmotionConfused  Emotion = "困惑"
	EmotionCalm      Emotion = "平静"
	EmotionAfraid    Emotion = "害怕"
	EmotionSurprised Emotion = "惊喜"
)

var allEmotions = []Emotion{EmotionHappy, EmotionTense, EmotionConfused, EmotionCalm, EmotionAfraid, EmotionSurprised}

func Emotions() []Emotion { return append([]Emotion(nil), allEmotions...) }

func ParseEmotion(
	value string,
) (Emotion, error) {
	for i := range allEmotions {
		emotion := allEmotions[i]
		if string(emotion) == value {
			return emotion, nil
		}
	}
	return "", fmt.Errorf("不支持的情绪: %s", value)
}

type ElementKind string

const (
	ElementPerson ElementKind = "人物"
	ElementAnimal ElementKind = "动物"
	ElementPlace  ElementKind = "地点"
	ElementObject ElementKind = "物品"
	ElementAction ElementKind = "动作"
	ElementColor  ElementKind = "颜色"
	ElementNature ElementKind = "自然"
)

var allElementKinds = []ElementKind{ElementPerson, ElementAnimal, ElementPlace, ElementObject, ElementAction, ElementColor, ElementNature}

func ElementKinds() []ElementKind { return append([]ElementKind(nil), allElementKinds...) }

func ParseElementKind(
	value string,
) (ElementKind, error) {
	for i := range allElementKinds {
		kind := allElementKinds[i]
		if string(kind) == value {
			return kind, nil
		}
	}
	return "", fmt.Errorf("不支持的元素分类: %s", value)
}

type ReportState string

const (
	ReportPending ReportState = "待刷新"
	ReportReady   ReportState = "已生成"
	ReportFailed  ReportState = "生成失败"
)

type AnalysisStatus string

const (
	AnalysisQueued  AnalysisStatus = "排队中"
	AnalysisRunning AnalysisStatus = "分析中"
	AnalysisDone    AnalysisStatus = "已完成"
	AnalysisFailed  AnalysisStatus = "失败"
)
