package service

import (
	"testing"
	"time"

	"dream117/internal/analysis"
	"dream117/internal/domain"
	"dream117/internal/store"
)

// TestAnalysisWordCloudHTTPChainOrdering verifies the HTTP output chain: the
// wordcloud path of AnalysisService.Analyze must return exactly the weight-
// ordered slice produced by analysis.WordCloud, so the API, the analysis
// report and the sentence summary all consume one ordering.
func TestAnalysisWordCloudHTTPChainOrdering(t *testing.T) {
	dir := t.TempDir()
	data, err := store.Open(dir + "/dreams.json")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	userID := "user-1"
	now := time.Now().UTC()
	// Title core term 时钟 appears in two titles; body term 大雨 appears many
	// times across dreams.
	dreams := []domain.Dream{
		{ID: "d1", UserID: userID, Title: "时钟响了", Content: "大雨 大雨 大雨", DreamDate: now, Clarity: 6, Emotion: domain.EmotionCalm, SleepHours: 7},
		{ID: "d2", UserID: userID, Title: "时钟停了", Content: "大雨 大雨 大雨 大雨", DreamDate: now, Clarity: 6, Emotion: domain.EmotionCalm, SleepHours: 7},
		{ID: "d3", UserID: userID, Title: "桥上", Content: "大雨 大雨", DreamDate: now, Clarity: 6, Emotion: domain.EmotionCalm, SleepHours: 7},
		{ID: "d4", UserID: userID, Title: "巷口", Content: "大雨 大雨", DreamDate: now, Clarity: 6, Emotion: domain.EmotionCalm, SleepHours: 7},
	}
	for i := range dreams {
		if err := data.SaveDream(dreams[i]); err != nil {
			t.Fatalf("save dream: %v", err)
		}
	}

	svc := NewAnalysisService(data)
	window := analysis.Window{
		From:  now.AddDate(0, 0, -1),
		To:    now.AddDate(0, 0, 1),
		Label: "30天",
	}

	result := svc.Analyze(userID, "wordcloud", window)
	stats, ok := result.([]domain.WordStat)
	if !ok {
		t.Fatalf("wordcloud result is %T, want []domain.WordStat", result)
	}

	items := data.ListDreams(userID, window.From, window.To, "")
	expected := analysis.WordCloud(items, 40)

	if len(stats) != len(expected) {
		t.Fatalf("length mismatch: http=%d expected=%d", len(stats), len(expected))
	}
	for i := range expected {
		if stats[i] != expected[i] {
			t.Fatalf("HTTP chain diverges at %d: %+v vs %+v", i, stats[i], expected[i])
		}
	}

	// The title term must lead, proving the HTTP response preserved the
	// weight ordering instead of a count-only sort.
	if len(stats) == 0 || stats[0].Word != "时钟" {
		t.Fatalf("expected title term 时钟 first, got %+v", stats[:min(3, len(stats))])
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
