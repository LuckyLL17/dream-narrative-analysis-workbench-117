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

func newAnalysisStore(t *testing.T) *store.Store {
	t.Helper()
	dir := t.TempDir()
	s, err := store.Open(filepath.Join(dir, "dreams.json"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	return s
}

func saveSimilarDreams(t *testing.T, s *store.Store, userID string) (target, other domain.Dream) {
	t.Helper()
	base := time.Date(2026, 8, 1, 7, 0, 0, 0, time.UTC)
	// Target dream shares a tag and a theme with `other` so the two produce a
	// non-zero similarity score.
	target = domain.Dream{
		ID:        ids.New("d"),
		UserID:    userID,
		Title:     "海边蓝色车站",
		Content:   "梦里我在海边蓝色车站等车",
		DreamDate: base,
		Tags:      []string{"车站", "海"},
		Themes:    []domain.ThemeHit{{Name: "被追赶", Evidence: []string{"追赶"}}},
		Emotion:   domain.EmotionCalm,
		Clarity:   6,
	}
	other = domain.Dream{
		ID:        ids.New("d"),
		UserID:    userID,
		Title:     "海边蓝色车站",
		Content:   "梦里我在海边蓝色车站等车",
		DreamDate: base.AddDate(0, 0, 1),
		Tags:      []string{"车站", "海"},
		Themes:    []domain.ThemeHit{{Name: "被追赶", Evidence: []string{"追赶"}}},
		Emotion:   domain.EmotionCalm,
		Clarity:   6,
	}
	if err := s.SaveDream(target); err != nil {
		t.Fatalf("save target: %v", err)
	}
	if err := s.SaveDream(other); err != nil {
		t.Fatalf("save other: %v", err)
	}
	return target, other
}

func TestSimilarReturnsLabelAndBandConsistentWithAnalysisLayer(t *testing.T) {
	s := newAnalysisStore(t)
	target, _ := saveSimilarDreams(t, s, "user-1")
	svc := NewAnalysisService(s)

	result, err := svc.Similar("user-1", target.ID, 0)
	if err != nil {
		t.Fatalf("Similar: %v", err)
	}
	if len(result.Items) != 1 {
		t.Fatalf("expected one similar dream, got %d", len(result.Items))
	}
	item := result.Items[0]
	if item.Label != analysis.SimilarityLabel(item.Score) {
		t.Fatalf("service label %q != analysis label %q for score %.3f", item.Label, analysis.SimilarityLabel(item.Score), item.Score)
	}
	if item.Band != analysis.SimilarityBandName(item.Score) {
		t.Fatalf("service band %q != analysis band %q for score %.3f", item.Band, analysis.SimilarityBandName(item.Score), item.Score)
	}
}

func TestSimilarDefaultLimitStableAcrossZeroAndNegative(t *testing.T) {
	s := newAnalysisStore(t)
	target, _ := saveSimilarDreams(t, s, "user-1")
	svc := NewAnalysisService(s)

	// Zero and negative limits must both normalise to the single default so
	// the HTTP route and the service cannot disagree on page size.
	for _, limit := range []int{0, -3} {
		result, err := svc.Similar("user-1", target.ID, limit)
		if err != nil {
			t.Fatalf("Similar(%d): %v", limit, err)
		}
		if result.Limit != analysis.DefaultSimilarLimit {
			t.Fatalf("limit=%d: expected default %d, got %d", limit, analysis.DefaultSimilarLimit, result.Limit)
		}
	}
}

func TestSimilarRespectsExplicitLimit(t *testing.T) {
	s := newAnalysisStore(t)
	target, _ := saveSimilarDreams(t, s, "user-1")
	// Add extra candidates so the limit is exercised.
	for i := 0; i < 3; i++ {
		extra := domain.Dream{
			ID:        ids.New("d"),
			UserID:    "user-1",
			Title:     "海边蓝色车站",
			Content:   "梦里我在海边蓝色车站等车",
			DreamDate: time.Date(2026, 8, 1, 7, 0, 0, 0, time.UTC).AddDate(0, 0, 2+i),
			Tags:      []string{"车站", "海"},
			Themes:    []domain.ThemeHit{{Name: "被追赶", Evidence: []string{"追赶"}}},
			Emotion:   domain.EmotionCalm,
			Clarity:   6,
		}
		if err := s.SaveDream(extra); err != nil {
			t.Fatalf("save extra: %v", err)
		}
	}
	svc := NewAnalysisService(s)

	result, err := svc.Similar("user-1", target.ID, 2)
	if err != nil {
		t.Fatalf("Similar: %v", err)
	}
	if len(result.Items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(result.Items))
	}
	if result.Limit != 2 {
		t.Fatalf("expected reported limit 2, got %d", result.Limit)
	}
}

func TestSimilarUnknownDreamReturnsNotFound(t *testing.T) {
	s := newAnalysisStore(t)
	svc := NewAnalysisService(s)
	if _, err := svc.Similar("user-1", "missing", 5); err != domain.ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestSimilarExcludesOtherUsersDreams(t *testing.T) {
	s := newAnalysisStore(t)
	target, _ := saveSimilarDreams(t, s, "user-1")
	// A second user's dream sharing every feature must not appear.
	otherUser := domain.Dream{
		ID:        ids.New("d"),
		UserID:    "user-2",
		Title:     "海边蓝色车站",
		Content:   "梦里我在海边蓝色车站等车",
		DreamDate: time.Date(2026, 8, 1, 7, 0, 0, 0, time.UTC),
		Tags:      []string{"车站", "海"},
		Themes:    []domain.ThemeHit{{Name: "被追赶", Evidence: []string{"追赶"}}},
		Emotion:   domain.EmotionCalm,
		Clarity:   6,
	}
	if err := s.SaveDream(otherUser); err != nil {
		t.Fatalf("save other user dream: %v", err)
	}
	svc := NewAnalysisService(s)
	result, err := svc.Similar("user-1", target.ID, 5)
	if err != nil {
		t.Fatalf("Similar: %v", err)
	}
	for i := range result.Items {
		if result.Items[i].Dream.UserID != "user-1" {
			t.Fatalf("similar dream leaked across users: %s", result.Items[i].Dream.UserID)
		}
	}
}
