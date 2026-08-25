package verification

import (
	"dream117/internal/analysis"
	"dream117/internal/domain"
	"testing"
)

func TestBug021Themepairsareunordered(t *testing.T) {
	a := domain.Dream{ID: "a", Emotion: domain.EmotionTense, Themes: []domain.ThemeHit{{Name: "被追赶"}, {Name: "迷路"}}}
	b := domain.Dream{ID: "b", Emotion: domain.EmotionCalm, Themes: []domain.ThemeHit{{Name: "迷路"}, {Name: "被追赶"}}}
	pairs := analysis.ThemeCooccurrence([]domain.Dream{a, b}, 10)
	if len(pairs) != 1 {
		t.Fatalf("reverse theme order produced duplicate pairs: %+v", pairs)
	}
	if pairs[0].Count != 2 {
		t.Fatalf("pair count=%d want=2", pairs[0].Count)
	}
}

func TestBug021RegressionHealth(t *testing.T) {
	if got := analysis.SimilarityLabel(0.7); got == "" {
		t.Fatal("empty similarity label")
	}
}
