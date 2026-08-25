package analysis

import (
	"testing"

	"dream117/internal/domain"
)

// regressionDreams reproduces the reported scenario: two dreams whose titles
// share a core term (时钟) while a body term (大雨) appears far more often.
// Per the product rule, the title's core term must outrank the higher-
// frequency body term in every word output.
func regressionDreams() []domain.Dream {
	return []domain.Dream{
		{ID: "d1", UserID: "u", Title: "时钟响了", Content: "大雨 大雨 大雨"},
		{ID: "d2", UserID: "u", Title: "时钟停了", Content: "大雨 大雨 大雨 大雨"},
		{ID: "d3", UserID: "u", Title: "桥上", Content: "大雨 大雨"},
		{ID: "d4", UserID: "u", Title: "巷口", Content: "大雨 大雨"},
		// Total occurrences across dreams: 时钟 -> 2 (title), 大雨 -> 7 (body).
	}
}

func TestWordCloudTitleTermOutranksBodyTerm(t *testing.T) {
	items := regressionDreams()
	wc := WordCloud(items, 40)

	clockIdx, rainIdx := -1, -1
	for i, w := range wc {
		switch w.Word {
		case "时钟":
			clockIdx = i
		case "大雨":
			rainIdx = i
		}
	}
	if clockIdx == -1 {
		t.Fatalf("title term 时钟 missing from wordcloud: %+v", wc)
	}
	if rainIdx == -1 {
		t.Fatalf("body term 大雨 missing from wordcloud: %+v", wc)
	}
	if clockIdx > rainIdx {
		t.Fatalf("title term 时钟 (index %d) must precede body term 大雨 (index %d); order=%v",
			clockIdx, rainIdx, words(wc))
	}

	clock := wc[clockIdx]
	rain := wc[rainIdx]
	if clock.Weight <= rain.Weight {
		t.Fatalf("title term weight %d must exceed body term weight %d", clock.Weight, rain.Weight)
	}
	if clock.Count > rain.Count {
		t.Fatalf("test fixture broken: expected body term to have higher count, got clock=%d rain=%d",
			clock.Count, rain.Count)
	}
}

// TestWordCloudRankedWordsSentenceShareOrdering ensures the three output
// surfaces derive their ordering from the same unified weight and never
// diverge. This is the core consistency contract described in the bug report.
func TestWordCloudRankedWordsSentenceShareOrdering(t *testing.T) {
	items := regressionDreams()

	ranked := RankedWords(items, 40)
	cloud := WordCloud(items, 40)

	if len(ranked) != len(cloud) {
		t.Fatalf("RankedWords and WordCloud lengths differ: %d vs %d", len(ranked), len(cloud))
	}
	for i := range ranked {
		if ranked[i] != cloud[i] {
			t.Fatalf("WordCloud diverges from RankedWords at %d: %+v vs %+v", i, ranked[i], cloud[i])
		}
	}

	// The sentence must lead with the title term (the highest-weight word),
	// proving it consumes the same ordering rather than a count-only one.
	sentence := WordCloudSentence(cloud)
	want := "时钟"
	r := []rune(sentence)
	if len(r) == 0 || string(r[0:len([]rune(want))]) != want {
		t.Fatalf("sentence should start with %q, got %q", want, truncate(sentence, 24))
	}
}

// TestWordCloudOrderingIsMonotonicByWeight confirms the final ordering is
// driven by the unified weight, not by raw body frequency.
func TestWordCloudOrderingIsMonotonicByWeight(t *testing.T) {
	items := regressionDreams()
	wc := WordCloud(items, 40)
	for i := 1; i < len(wc); i++ {
		if wc[i-1].Weight < wc[i].Weight {
			t.Fatalf("wordcloud not weight-sorted at %d: %+v then %+v", i-1, wc[i-1], wc[i])
		}
	}
}

func words(stats []domain.WordStat) []string {
	out := make([]string, 0, len(stats))
	for _, s := range stats {
		out = append(out, s.Word)
	}
	return out
}

func truncate(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n]) + "…"
}
