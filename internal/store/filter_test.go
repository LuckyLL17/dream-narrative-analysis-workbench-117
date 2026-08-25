package store

import (
	"path/filepath"
	"testing"
	"time"

	"dream117/internal/domain"
)

// seedPagination reproduces the reported scenario: three dreams on the
// selected end day for one user, the last one timestamped at the very last
// representable nanosecond of that day.
func seedPagination(t *testing.T) (*Store, string) {
	t.Helper()
	store, err := Open(filepath.Join(t.TempDir(), "dreams.json"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	userID := "user-pagination"
	must := func(d domain.Dream) {
		if err := store.SaveDream(d); err != nil {
			t.Fatalf("save dream: %v", err)
		}
	}
	day := time.Date(2026, 8, 2, 0, 0, 0, 0, time.UTC)
	content := "这是一段梦境描述用于测试"
	must(domain.Dream{ID: "d1", UserID: userID, Title: "分页一", Content: content, DreamDate: day.Add(10 * time.Hour), WakeTime: day.Add(11 * time.Hour), Clarity: 5, Emotion: domain.EmotionCalm, CreatedAt: day.Add(10 * time.Hour)})
	must(domain.Dream{ID: "d2", UserID: userID, Title: "分页二", Content: content, DreamDate: day.Add(11 * time.Hour), WakeTime: day.Add(12 * time.Hour), Clarity: 5, Emotion: domain.EmotionCalm, CreatedAt: day.Add(11 * time.Hour)})
	must(domain.Dream{ID: "d3", UserID: userID, Title: "分页三", Content: content, DreamDate: clockDayEnd(day), WakeTime: day.AddDate(0, 0, 1), Clarity: 5, Emotion: domain.EmotionCalm, CreatedAt: clockDayEnd(day).Add(-time.Second)})
	// An unrelated user's dream must never bleed into the result set.
	must(domain.Dream{ID: "dX", UserID: "user-other", Title: "分页一", Content: content, DreamDate: day.Add(10 * time.Hour), WakeTime: day.Add(11 * time.Hour), Clarity: 5, Emotion: domain.EmotionCalm, CreatedAt: day.Add(10 * time.Hour)})
	return store, userID
}

func clockDayEnd(value time.Time) time.Time {
	return value.Add(24*time.Hour - time.Nanosecond)
}

// TestSearchPagePaginationContract asserts the whole pagination contract for
// the reported case: total=3, total_pages=2, page 2 returns the chronologically
// oldest ("分页一") of the three matches after descending sort.
func TestSearchPagePaginationContract(t *testing.T) {
	store, userID := seedPagination(t)
	page := store.SearchPage(userID, domain.DreamFilter{
		From:     time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC),
		To:       time.Date(2026, 8, 2, 0, 0, 0, 0, time.UTC), // bare date → must be ceiled to end of day
		Query:    "分页",
		Page:     2,
		PageSize: 2,
	})

	if page.Total != 3 {
		t.Fatalf("expected total=3, got total=%d", page.Total)
	}
	if page.TotalPages != 2 {
		t.Fatalf("expected total_pages=2, got total_pages=%d", page.TotalPages)
	}
	if page.Page != 2 {
		t.Fatalf("expected echoed page=2, got page=%d", page.Page)
	}
	if page.PageSize != 2 {
		t.Fatalf("expected page_size=2, got page_size=%d", page.PageSize)
	}
	if got := len(page.Items); got != 1 {
		t.Fatalf("expected 1 item on page 2, got %d", got)
	}
	if got, want := page.Items[0].Title, "分页一"; got != want {
		t.Fatalf("expected page 2 to be %q after descending sort, got %q", want, got)
	}
	if page.HasNext {
		t.Fatalf("expected HasNext=false on the last page")
	}
	if !page.HasPrevious {
		t.Fatalf("expected HasPrevious=true on page 2")
	}
}

// TestSearchPageFirstPageFull asserts the off-by-one slice bug is fixed: page 1
// with page_size=2 over three matches must return two items, not one.
func TestSearchPageFirstPageFull(t *testing.T) {
	store, userID := seedPagination(t)
	page := store.SearchPage(userID, domain.DreamFilter{
		From:     time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC),
		To:       time.Date(2026, 8, 2, 0, 0, 0, 0, time.UTC),
		Query:    "分页",
		Page:     1,
		PageSize: 2,
	})
	if got := len(page.Items); got != 2 {
		t.Fatalf("expected page 1 to hold 2 items, got %d", got)
	}
	if page.Items[0].Title != "分页三" || page.Items[1].Title != "分页二" {
		t.Fatalf("expected descending order [分页三, 分页二], got %q then %q", page.Items[0].Title, page.Items[1].Title)
	}
	if !page.HasNext {
		t.Fatalf("expected HasNext=true on page 1 when a second page exists")
	}
	if page.HasPrevious {
		t.Fatalf("expected HasPrevious=false on page 1")
	}
}

// TestSearchPageEndOfDayInclusive asserts a record timestamped at the last
// nanosecond of the selected end day is kept (the date-bound regression).
func TestSearchPageEndOfDayInclusive(t *testing.T) {
	store, userID := seedPagination(t)
	page := store.SearchPage(userID, domain.DreamFilter{
		From:     time.Date(2026, 8, 2, 0, 0, 0, 0, time.UTC),
		To:       time.Date(2026, 8, 2, 0, 0, 0, 0, time.UTC),
		Query:    "分页三",
		Page:     1,
		PageSize: 10,
	})
	if page.Total != 1 {
		t.Fatalf("expected the end-of-day record to participate (total=1), got total=%d", page.Total)
	}
	if len(page.Items) != 1 || page.Items[0].Title != "分页三" {
		t.Fatalf("expected 分页三, got %+v", page.Items)
	}
}

// TestSearchPageOutOfBoundsIsEmpty asserts an out-of-range page echoes back the
// requested page with an empty slice and correct totals rather than panicking
// or substituting another page.
func TestSearchPageOutOfBoundsIsEmpty(t *testing.T) {
	store, userID := seedPagination(t)
	page := store.SearchPage(userID, domain.DreamFilter{
		From:     time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC),
		To:       time.Date(2026, 8, 2, 0, 0, 0, 0, time.UTC),
		Query:    "分页",
		Page:     5,
		PageSize: 2,
	})
	if page.Total != 3 || page.TotalPages != 2 {
		t.Fatalf("expected total=3 total_pages=2, got total=%d total_pages=%d", page.Total, page.TotalPages)
	}
	if page.Page != 5 {
		t.Fatalf("expected requested page=5 to echo back, got page=%d", page.Page)
	}
	if len(page.Items) != 0 {
		t.Fatalf("expected empty slice on out-of-range page, got %d items", len(page.Items))
	}
}

// TestSearchPageEmptyResult asserts an empty result still reports consistent
// metadata (total=0, total_pages=1) instead of "total=0 total_pages=0".
func TestSearchPageEmptyResult(t *testing.T) {
	store, userID := seedPagination(t)
	page := store.SearchPage(userID, domain.DreamFilter{
		Query:    "不存在的关键词",
		Page:     1,
		PageSize: 2,
	})
	if page.Total != 0 {
		t.Fatalf("expected total=0, got total=%d", page.Total)
	}
	if page.TotalPages != 1 {
		t.Fatalf("expected total_pages=1 for empty result, got total_pages=%d", page.TotalPages)
	}
	if len(page.Items) != 0 {
		t.Fatalf("expected empty items, got %d", len(page.Items))
	}
	if page.HasNext || page.HasPrevious {
		t.Fatalf("expected no next/previous on empty result, got has_next=%v has_previous=%v", page.HasNext, page.HasPrevious)
	}
}

// TestMatchesFilterBounds documents the inclusive [From, To] day contract that
// both bounds are honored and the upper bound is inclusive.
func TestMatchesFilterBounds(t *testing.T) {
	midnight := time.Date(2026, 8, 2, 0, 0, 0, 0, time.UTC)
	filter := domain.DreamFilter{From: midnight, To: clockDayEnd(midnight)}.Normalized()
	cases := map[string]struct {
		date time.Time
		want bool
	}{
		"start_of_day":       {midnight, true},
		"middle_of_day":      {midnight.Add(12 * time.Hour), true},
		"end_of_day_last_ns": {clockDayEnd(midnight), true},
		"day_before":         {midnight.Add(-time.Nanosecond), false},
		"day_after_first_ns": {clockDayEnd(midnight).Add(time.Nanosecond), false},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			d := domain.Dream{UserID: "u", Title: "分页", Content: "content here", DreamDate: tc.date, WakeTime: tc.date.Add(time.Hour), Clarity: 5, Emotion: domain.EmotionCalm}
			if got := matchesFilter(d, filter); got != tc.want {
				t.Fatalf("matchesFilter(%s)=%v, want %v", tc.date.Format(time.RFC3339Nano), got, tc.want)
			}
		})
	}
}
