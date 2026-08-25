package verification

import (
	"dream117/internal/analysis"
	"testing"
)

func TestBug027Similaritybandboundaries(t *testing.T) {
	cases := map[float64]string{0.4: "部分相似", 0.7: "强相似", 0.39: "轻微重合"}
	for score, want := range cases {
		if got := analysis.SimilarityLabel(score); got != want {
			t.Fatalf("score=%v label=%q want=%q", score, got, want)
		}
	}
}

func TestBug027RegressionHealth(t *testing.T) {
	if len(analysis.SimilarityBands()) != 3 {
		t.Fatal("similarity bands changed unexpectedly")
	}
}
