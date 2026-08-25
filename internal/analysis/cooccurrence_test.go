package analysis

import (
	"testing"
	"time"

	"dream117/internal/domain"
)

func dreamAt(date time.Time, emotion string, themeNames ...string) domain.Dream {
	themes := make([]domain.ThemeHit, 0, len(themeNames))
	for _, n := range themeNames {
		themes = append(themes, domain.ThemeHit{Name: n})
	}
	return domain.Dream{Emotion: domain.Emotion(emotion), Themes: themes, DreamDate: date}
}

func findPair(pairs []ThemePair, left, right string) (ThemePair, bool) {
	for _, p := range pairs {
		if (p.Left == left && p.Right == right) || (p.Left == right && p.Right == left) {
			return p, true
		}
	}
	return ThemePair{}, false
}

// Two dreams that describe the same themes in opposite orders must collapse to
// a single co-occurrence edge with the merged count.
func TestThemeCooccurrenceMergesReversedOrder(t *testing.T) {
	day := time.Date(2026, 8, 1, 6, 0, 0, 0, time.UTC)
	items := []domain.Dream{
		dreamAt(day, "害怕", "被追赶", "迷路"),
		dreamAt(day.AddDate(0, 0, 1), "害怕", "迷路", "被追赶"),
	}
	pairs := ThemeCooccurrence(items, 24)

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

// Count must reflect the number of dreams sharing the pair, not the number of
// orderings. Here three dreams all carry the same two themes in mixed orders.
func TestThemeCooccurrenceCountMatchesDreamCount(t *testing.T) {
	day := time.Date(2026, 8, 1, 6, 0, 0, 0, time.UTC)
	items := []domain.Dream{
		dreamAt(day, "紧张", "被追赶", "迷路"),
		dreamAt(day.AddDate(0, 0, 1), "害怕", "迷路", "被追赶"),
		dreamAt(day.AddDate(0, 0, 2), "紧张", "被追赶", "迷路"),
	}
	pairs := ThemeCooccurrence(items, 24)

	got, ok := findPair(pairs, "被追赶", "迷路")
	if !ok {
		t.Fatalf("missing 被追赶—迷路 edge in %#v", pairs)
	}
	if got.Count != 3 {
		t.Errorf("Count: want 3, got %d", got.Count)
	}
	// Each theme appears in all 3 dreams; union of theme scopes is 3; 3/3 = 1.
	if got.Jaccard != 1 {
		t.Errorf("Jaccard: want 1, got %v", got.Jaccard)
	}
	if got.SharedMood != "紧张" {
		t.Errorf("SharedMood: want 紧张 (2 vs 1), got %s", got.SharedMood)
	}
}

// Jaccard is co-occurrence over the union of the two themes' dream scopes.
// Theme A in 3 dreams, theme B in 2 dreams, both in 2 dreams -> union 3, 2/3.
func TestThemeCooccurrenceJaccardPartialOverlap(t *testing.T) {
	day := time.Date(2026, 8, 1, 6, 0, 0, 0, time.UTC)
	items := []domain.Dream{
		dreamAt(day, "害怕", "被追赶", "迷路"),
		dreamAt(day.AddDate(0, 0, 1), "害怕", "迷路", "被追赶"),
		dreamAt(day.AddDate(0, 0, 2), "害怕", "被追赶"),
	}
	pairs := ThemeCooccurrence(items, 24)

	got, ok := findPair(pairs, "被追赶", "迷路")
	if !ok {
		t.Fatalf("missing 被追赶—迷路 edge in %#v", pairs)
	}
	if got.Count != 2 {
		t.Errorf("Count: want 2, got %d", got.Count)
	}
	// 被追赶 appears in 3 dreams, 迷路 in 2, shared in 2 -> union 3 -> 2/3 ≈ 0.667.
	if got.Jaccard != 0.667 {
		t.Errorf("Jaccard: want 0.667, got %v", got.Jaccard)
	}
}

// A dream with only one theme produces no pairs.
func TestThemeCooccurrenceNoPairForSingleTheme(t *testing.T) {
	day := time.Date(2026, 8, 1, 6, 0, 0, 0, time.UTC)
	items := []domain.Dream{
		dreamAt(day, "平静", "飞行"),
	}
	pairs := ThemeCooccurrence(items, 24)
	if len(pairs) != 0 {
		t.Fatalf("expected no pairs, got %#v", pairs)
	}
}

// Within a single dream, repeated theme names must not inflate the pair count.
func TestThemeCooccurrenceDedupesRepeatedThemes(t *testing.T) {
	day := time.Date(2026, 8, 1, 6, 0, 0, 0, time.UTC)
	items := []domain.Dream{
		dreamAt(day, "害怕", "被追赶", "被追赶", "迷路", "迷路"),
	}
	pairs := ThemeCooccurrence(items, 24)
	got, ok := findPair(pairs, "被追赶", "迷路")
	if !ok {
		t.Fatalf("missing 被追赶—迷路 edge in %#v", pairs)
	}
	if got.Count != 1 {
		t.Errorf("Count: want 1 (one dream), got %d", got.Count)
	}
}

// SharedMood is computed over both themes' mood distributions, so it must
// reflect the merged counts rather than a single ordering's slice.
func TestThemeCooccurrenceSharedMoodMergesBothScopes(t *testing.T) {
	day := time.Date(2026, 8, 1, 6, 0, 0, 0, time.UTC)
	items := []domain.Dream{
		dreamAt(day, "害怕", "被追赶", "迷路"),
		dreamAt(day.AddDate(0, 0, 1), "紧张", "被追赶", "迷路"),
		dreamAt(day.AddDate(0, 0, 2), "害怕", "被追赶", "迷路"),
	}
	pairs := ThemeCooccurrence(items, 24)
	got, ok := findPair(pairs, "被追赶", "迷路")
	if !ok {
		t.Fatalf("missing edge in %#v", pairs)
	}
	if got.SharedMood != "害怕" {
		t.Errorf("SharedMood: want 害怕 (2 vs 1), got %s", got.SharedMood)
	}
}

// Three themes in one dream yield three unordered pairs, each Count=1.
func TestThemeCooccurrenceThreeThemesFormThreeEdges(t *testing.T) {
	day := time.Date(2026, 8, 1, 6, 0, 0, 0, time.UTC)
	items := []domain.Dream{
		dreamAt(day, "害怕", "被追赶", "迷路", "坠落"),
	}
	pairs := ThemeCooccurrence(items, 24)
	if len(pairs) != 3 {
		t.Fatalf("expected 3 edges, got %d: %#v", len(pairs), pairs)
	}
	for _, want := range [][2]string{{"被追赶", "迷路"}, {"被追赶", "坠落"}, {"迷路", "坠落"}} {
		if _, ok := findPair(pairs, want[0], want[1]); !ok {
			t.Errorf("missing edge %s—%s in %#v", want[0], want[1], pairs)
		}
	}
}

// The result is deterministic regardless of which dream carries which order.
func TestThemeCooccurrenceStableAcrossInputOrdering(t *testing.T) {
	day := time.Date(2026, 8, 1, 6, 0, 0, 0, time.UTC)
	dreamsAB := []domain.Dream{
		dreamAt(day, "害怕", "被追赶", "迷路"),
		dreamAt(day.AddDate(0, 0, 1), "害怕", "迷路", "被追赶"),
	}
	dreamsBA := []domain.Dream{
		dreamAt(day, "害怕", "迷路", "被追赶"),
		dreamAt(day.AddDate(0, 0, 1), "害怕", "被追赶", "迷路"),
	}
	first := ThemeCooccurrence(dreamsAB, 24)
	second := ThemeCooccurrence(dreamsBA, 24)
	if len(first) != 1 || len(second) != 1 {
		t.Fatalf("expected 1 edge each, got %d and %d", len(first), len(second))
	}
	if first[0] != second[0] {
		t.Errorf("order-dependent result: %#v vs %#v", first[0], second[0])
	}
}

// pairKey is the unit the order-independence guarantee rests on.
func TestPairKeyOrderIndependent(t *testing.T) {
	if pairKey("被追赶", "迷路") != pairKey("迷路", "被追赶") {
		t.Fatal("pairKey must be identical for the two orderings of a pair")
	}
	if got := pairKey("飞行", "飞行"); got != "飞行\x00飞行" {
		t.Errorf("pairKey equal-name: want 飞行\\x00飞行, got %q", got)
	}
}
