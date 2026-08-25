package analysis

import (
	"encoding/json"
	"math"
	"strconv"
	"strings"
	"testing"
	"time"

	"dream117/internal/domain"
	"dream117/pkg/jsonutil"
)

func TestSimilarityLabelBoundaries(t *testing.T) {
	cases := []struct {
		name  string
		score float64
		want  string
	}{
		{"zero is weak", 0, "轻微重合"},
		{"just under weak top", 0.39, "轻微重合"},
		{"exact medium boundary is medium", 0.4, "部分相似"},
		{"mid medium", 0.5, "部分相似"},
		{"just under strong boundary stays medium", 0.699, "部分相似"},
		{"exact strong boundary is strong", 0.7, "强相似"},
		{"above old wrong strong boundary", 0.75, "强相似"},
		{"max score is strong", 1, "强相似"},
		{"above max still strong", 1.2, "强相似"},
		{"negative falls back to weak", -0.01, "轻微重合"},
	}
	for i := range cases {
		c := cases[i]
		t.Run(c.name, func(t *testing.T) {
			if got := SimilarityLabel(c.score); got != c.want {
				t.Fatalf("SimilarityLabel(%.3f) = %q, want %q", c.score, got, c.want)
			}
		})
	}
}

func TestSimilarityBandNameMirrorsLabel(t *testing.T) {
	pairs := map[float64]string{
		0:     "weak",
		0.39:  "weak",
		0.4:   "medium",
		0.7:   "strong",
		1:     "strong",
		-0.5:  "weak",
	}
	for score, want := range pairs {
		name := SimilarityBandName(score)
		if name != want {
			t.Fatalf("SimilarityBandName(%.3f) = %q, want %q", score, name, want)
		}
		// The label and band must resolve to the same band entry.
		band := bandByName(name)
		if band.Label != SimilarityLabel(score) {
			t.Fatalf("score %.3f: band %q label %q != SimilarityLabel %q", score, name, band.Label, SimilarityLabel(score))
		}
	}
}

func bandByName(name string) SimilarityBand {
	for i := range SimilarityBands() {
		b := SimilarityBands()[i]
		if b.Name == name {
			return b
		}
	}
	return SimilarityBand{}
}

func TestSimilarityBandsAreContiguousLeftClosed(t *testing.T) {
	bands := SimilarityBands()
	if len(bands) != 3 {
		t.Fatalf("expected three bands, got %d", len(bands))
	}
	// weak starts at 0; each band's Max equals the next band's Min (no gap,
	// no overlap, left-closed right-open).
	if bands[0].Min != 0 {
		t.Fatalf("weak should start at 0, got %v", bands[0].Min)
	}
	for i := 0; i < len(bands)-1; i++ {
		if bands[i].Max != bands[i+1].Min {
			t.Fatalf("band %s max %v != band %s min %v", bands[i].Name, bands[i].Max, bands[i+1].Name, bands[i+1].Min)
		}
		if bands[i].Max <= bands[i].Min {
			t.Fatalf("band %s is not left-closed right-open: [%v, %v)", bands[i].Name, bands[i].Min, bands[i].Max)
		}
	}
	// No gap between medium and strong: the historic bug had strong start at 0.75.
	if bands[2].Min != 0.7 {
		t.Fatalf("strong must start at 0.7 to avoid the 0.7 gap, got %v", bands[2].Min)
	}
}

// dreamForSimilarity builds a dream whose feature overlap with another dream is
// fully determined by tags and themes so a boundary score can be asserted.
func dreamForSimilarity(id, title string, tags, themes []string) domain.Dream {
	hits := make([]domain.ThemeHit, 0, len(themes))
	for i := range themes {
		hits = append(hits, domain.ThemeHit{Name: themes[i], Evidence: []string{themes[i]}})
	}
	return domain.Dream{
		ID:        id,
		UserID:    "user-1",
		Title:     title,
		Content:   title,
		DreamDate: time.Date(2026, 8, 1, 7, 0, 0, 0, time.UTC),
		Tags:      tags,
		Themes:    hits,
		Emotion:   domain.EmotionCalm,
		Clarity:   6,
	}
}

func TestSimilarDreamsLabelsAndBandsMatchScore(t *testing.T) {
	target := dreamForSimilarity("d1", "海边蓝色车站", []string{"车站", "海"}, []string{"被追赶"})
	// Two candidates with the same features as the target produce a non-zero
	// score; their label/band must come from the same source as the score.
	same := dreamForSimilarity("d2", "海边蓝色车站", []string{"车站", "海"}, []string{"被追赶"})
	res := SimilarDreams(target, []domain.Dream{same}, 5)
	if len(res.Items) != 1 {
		t.Fatalf("expected one similar dream, got %d", len(res.Items))
	}
	item := res.Items[0]
	if item.Score <= 0 {
		t.Fatalf("expected positive score, got %v", item.Score)
	}
	wantLabel := SimilarityLabel(item.Score)
	if item.Label != wantLabel {
		t.Fatalf("item.Label %q != SimilarityLabel(score) %q", item.Label, wantLabel)
	}
	if item.Band != SimilarityBandName(item.Score) {
		t.Fatalf("item.Band %q != SimilarityBandName(score) %q", item.Band, SimilarityBandName(item.Score))
	}
}

func TestSimilarDreamsResultReportsBandsAndLimit(t *testing.T) {
	target := dreamForSimilarity("d1", "标题", []string{"a"}, []string{"被追赶"})
	cands := make([]domain.Dream, 0, 12)
	for i := 0; i < 12; i++ {
		cands = append(cands, dreamForSimilarity("d"+strconv.Itoa(i+2), "标题", []string{"a"}, []string{"被追赶"}))
	}
	res := SimilarDreams(target, cands, 0)
	if res.Limit != DefaultSimilarLimit {
		t.Fatalf("expected default limit %d, got %d", DefaultSimilarLimit, res.Limit)
	}
	if len(res.Items) != DefaultSimilarLimit {
		t.Fatalf("expected %d items, got %d", DefaultSimilarLimit, len(res.Items))
	}
	if len(res.Bands) != 3 {
		t.Fatalf("expected bands overview with 3 bands, got %d", len(res.Bands))
	}
	// Every reported band must carry a non-empty label and note so the
	// overview segment description is never empty.
	for i := range res.Bands {
		if res.Bands[i].Label == "" || res.Bands[i].Note == "" {
			t.Fatalf("band %s missing label/note", res.Bands[i].Name)
		}
	}
}

func TestSimilarDreamsRespectsExplicitLimit(t *testing.T) {
	target := dreamForSimilarity("d1", "标题", []string{"a"}, []string{"被追赶"})
	cands := make([]domain.Dream, 0, 4)
	for i := 0; i < 4; i++ {
		cands = append(cands, dreamForSimilarity("d"+strconv.Itoa(i+2), "标题", []string{"a"}, []string{"被追赶"}))
	}
	res := SimilarDreams(target, cands, 2)
	if len(res.Items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(res.Items))
	}
	if res.Limit != 2 {
		t.Fatalf("expected limit 2, got %d", res.Limit)
	}
}

func TestSimilarDreamsExcludesTarget(t *testing.T) {
	target := dreamForSimilarity("d1", "标题", []string{"a"}, []string{"被追赶"})
	// Candidate list that includes the target itself; it must be skipped.
	res := SimilarDreams(target, []domain.Dream{target}, 5)
	if len(res.Items) != 0 {
		t.Fatalf("target should be excluded, got %d items", len(res.Items))
	}
}

func TestSimilarityLabelNaNDoesNotCrash(t *testing.T) {
	// NaN must not panic; it falls through to the weak fallback.
	got := SimilarityLabel(math.NaN())
	if got != "轻微重合" {
		t.Fatalf("NaN should fall back to 轻微重合, got %q", got)
	}
}

func TestSimilarDreamsResultSerializesWithLabelAndBands(t *testing.T) {
	target := dreamForSimilarity("d1", "标题", []string{"a"}, []string{"被追赶"})
	other := dreamForSimilarity("d2", "标题", []string{"a"}, []string{"被追赶"})
	res := SimilarDreams(target, []domain.Dream{other}, 5)

	data, err := jsonutil.Marshal(res)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	body := string(data)
	// The API contract: each item exposes its band/label and the response
	// carries the shared bands overview and the effective limit.
	for _, key := range []string{`"label"`, `"band"`, `"bands"`, `"limit"`, `"note"`, `"min"`, `"max"`} {
		if !strings.Contains(body, key) {
			t.Fatalf("serialized response missing %s: %s", key, body)
		}
	}
	// Round-trip back to confirm the snake_case payload decodes cleanly.
	var decoded SimilarDreamsResult
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(decoded.Items) != len(res.Items) || len(decoded.Bands) != len(res.Bands) {
		t.Fatalf("round-trip mismatch: items %d/%d bands %d/%d", len(decoded.Items), len(res.Items), len(decoded.Bands), len(res.Bands))
	}
	if decoded.Limit != res.Limit {
		t.Fatalf("round-trip limit mismatch: %d/%d", decoded.Limit, res.Limit)
	}
}
