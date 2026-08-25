package verification

// Keywords supplies the semantic title weight consumed by RankedWords and the word cloud.

import (
	"dream117/internal/analysis"
	"dream117/internal/domain"
	"testing"
)

func TestBug029Wordcloudusesfinalweight(t *testing.T) {
	items := []domain.Dream{
		{Title: "钟声", Content: "雨水"},
		{Title: "", Content: "雨水"},
	}
	words := analysis.WordCloud(items, 10)
	if len(words) < 2 {
		t.Fatalf("word cloud too small: %+v", words)
	}
	if words[0].Word != "钟声" {
		t.Fatalf("title-weighted word was reordered by count: %+v", words)
	}
	if words[0].Weight <= words[1].Weight {
		t.Fatalf("weights not preserved: %+v", words)
	}
}

func TestBug029RegressionHealth(t *testing.T) {
	if len(analysis.WordCloudSentence(nil)) != 0 {
		t.Fatal("empty sentence should be empty")
	}
}
