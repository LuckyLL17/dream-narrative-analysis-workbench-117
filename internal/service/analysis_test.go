package service

import (
	"path/filepath"
	"testing"
	"time"

	"dream117/internal/analysis"
	"dream117/internal/domain"
	"dream117/internal/store"
	"dream117/pkg/ids"
)

func newTestStore(t *testing.T) *store.Store {
	t.Helper()
	dir := t.TempDir()
	data, err := store.Open(filepath.Join(dir, "dreams.json"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	return data
}

func seedDream(s *store.Store, userID string, date time.Time, emotion string, themeNames ...string) {
	hits := make([]domain.ThemeHit, 0, len(themeNames))
	for _, n := range themeNames {
		hits = append(hits, domain.ThemeHit{Name: n})
	}
	if err := s.SaveDream(domain.Dream{ID: ids.New("dream"), UserID: userID, DreamDate: date, Emotion: domain.Emotion(emotion), Themes: hits}); err != nil {
		panic(err)
	}
}

// Regression for the report that two dreams carrying the same themes in
// opposite orders produced two reversed Count=1 edges. The full service ->
// analysis chain must return a single merged edge with Count=2.
func TestAnalysisServiceThemePairsMergesOrder(t *testing.T) {
	store := newTestStore(t)
	const userID = "u1"
	day := time.Date(2026, 8, 1, 6, 0, 0, 0, time.UTC)
	seedDream(store, userID, day, "害怕", "被追赶", "迷路")
	seedDream(store, userID, day.AddDate(0, 0, 1), "害怕", "迷路", "被追赶")

	svc := NewAnalysisService(store)
	window := analysis.Window{
		From:  day.AddDate(0, 0, -1),
		To:    day.AddDate(0, 0, 2),
		Label: "测试窗口",
	}

	result := svc.Analyze(userID, "theme-pairs", window)
	pairs, ok := result.([]analysis.ThemePair)
	if !ok {
		t.Fatalf("theme-pairs returned %T, want []analysis.ThemePair", result)
	}
	if len(pairs) != 1 {
		t.Fatalf("expected a single merged edge, got %d: %#v", len(pairs), pairs)
	}
	got := pairs[0]
	if got.Count != 2 {
		t.Errorf("Count: want 2, got %d", got.Count)
	}
	if got.Left != "被追赶" || got.Right != "迷路" {
		t.Errorf("edge endpoints: want 被追赶/迷路, got %s/%s", got.Left, got.Right)
	}
	if got.SharedMood != "害怕" {
		t.Errorf("SharedMood: want 害怕, got %s", got.SharedMood)
	}
}

// The same dream data feeds overview, theme pairs and timeline. Assert that
// overview theme stats and theme pairs agree on per-theme counts even when
// themes appear in different orders across dreams.
func TestAnalysisServiceOverviewAndPairsShareData(t *testing.T) {
	store := newTestStore(t)
	const userID = "u1"
	day := time.Date(2026, 8, 1, 6, 0, 0, 0, time.UTC)
	seedDream(store, userID, day, "害怕", "被追赶", "迷路")
	seedDream(store, userID, day.AddDate(0, 0, 1), "紧张", "迷路", "被追赶")
	seedDream(store, userID, day.AddDate(0, 0, 2), "害怕", "被追赶")

	svc := NewAnalysisService(store)
	window := analysis.Window{
		From:  day.AddDate(0, 0, -1),
		To:    day.AddDate(0, 0, 3),
		Label: "测试窗口",
	}

	overviewRaw := svc.Analyze(userID, "overview", window)
	overview, ok := overviewRaw.(domain.Overview)
	if !ok {
		t.Fatalf("overview returned %T, want domain.Overview", overviewRaw)
	}
	// 被追赶 appears in all 3 dreams, 迷路 in 2.
	var themeCounts = map[string]int{}
	for _, ts := range overview.TopThemes {
		themeCounts[ts.Name] = ts.Count
	}
	if themeCounts["被追赶"] != 3 {
		t.Errorf("被追赶 overview count: want 3, got %d", themeCounts["被追赶"])
	}
	if themeCounts["迷路"] != 2 {
		t.Errorf("迷路 overview count: want 2, got %d", themeCounts["迷路"])
	}

	pairsRaw := svc.Analyze(userID, "theme-pairs", window)
	pairs, ok := pairsRaw.([]analysis.ThemePair)
	if !ok {
		t.Fatalf("theme-pairs returned %T, want []analysis.ThemePair", pairsRaw)
	}
	var pairCount int
	for _, p := range pairs {
		if (p.Left == "被追赶" && p.Right == "迷路") || (p.Left == "迷路" && p.Right == "被追赶") {
			pairCount = p.Count
		}
	}
	if pairCount != 2 {
		t.Errorf("theme-pairs 被追赶—迷路 count: want 2 (matches 迷路 overview scope), got %d", pairCount)
	}
}
