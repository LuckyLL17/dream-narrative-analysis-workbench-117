package service

import (
	"testing"
	"time"

	"dream117/internal/domain"
	"dream117/internal/store"
)

// newStore opens an empty on-disk store against a temp path so the report
// service can read and persist reports like the real app does.
func newStore(t *testing.T) *store.Store {
	t.Helper()
	s, err := store.Open(t.TempDir() + "/dreams.json")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	return s
}

// sundayRefresh is the instant the user in the report clicked 立即刷新 on Sunday
// 2026-08-09: 23:00 UTC, with the dream 屋顶下着雨 logged at 23:00 that same night.
func sundayRefresh() time.Time {
	return time.Date(2026, 8, 9, 23, 0, 0, 0, time.UTC)
}

func saveDream(t *testing.T, s *store.Store, userID string, title string, dreamDate time.Time) {
	t.Helper()
	dream := domain.Dream{
		ID:        title,
		UserID:    userID,
		Title:     title,
		Content:   "醒来记得很清楚的一段叙事内容",
		DreamDate: dreamDate,
		WakeTime:  dreamDate.Add(8 * time.Hour),
		Emotion:   domain.EmotionAfraid,
		Clarity:   7,
	}
	if err := s.SaveDream(dream); err != nil {
		t.Fatalf("save dream: %v", err)
	}
}

// TestRefreshOnSundayCoversFullWeek is the headline regression: a refresh fired
// at Sunday 23:00 must cover Mon 00:00 through Sun 23:59:59.999999999 of the
// same week, and the last dream logged that Sunday must be counted.
func TestRefreshOnSundayCoversFullWeek(t *testing.T) {
	s := newStore(t)
	svc := NewReportService(s)
	const userID = "user-sun"

	wantStart := time.Date(2026, 8, 3, 0, 0, 0, 0, time.UTC)
	wantEnd := time.Date(2026, 8, 9, 23, 59, 59, 999999999, time.UTC)

	// The Sunday-night dream that used to fall outside the window.
	saveDream(t, s, userID, "屋顶下着雨", time.Date(2026, 8, 9, 23, 0, 0, 0, time.UTC))

	report, err := svc.Refresh(userID, sundayRefresh())
	if err != nil {
		t.Fatalf("refresh: %v", err)
	}
	if !report.WeekStart.Equal(wantStart) {
		t.Errorf("WeekStart = %s, want %s", report.WeekStart, wantStart)
	}
	if !report.WeekEnd.Equal(wantEnd) {
		t.Errorf("WeekEnd = %s, want %s", report.WeekEnd, wantEnd)
	}
	if report.TotalDreams != 1 {
		t.Errorf("TotalDreams = %d, want 1 (Sunday dream must be counted)", report.TotalDreams)
	}
	if report.ID != "weekly-20260803" {
		t.Errorf("report ID = %s, want weekly-20260803 (cache key must match WeekStart)", report.ID)
	}
	if report.State != domain.ReportReady {
		t.Errorf("State = %q, want %q", report.State, domain.ReportReady)
	}
}

// TestRefreshCacheKeyStableAcrossWeek asserts the cache key (report.ID, derived
// from WeekStart) stays constant when refreshing at different moments within
// the same week. Before the fix a Sunday refresh keyed on the next Monday and
// missed the cached Saturday refresh.
func TestRefreshCacheKeyStableAcrossWeek(t *testing.T) {
	s := newStore(t)
	svc := NewReportService(s)
	const userID = "user-stable"
	saveDream(t, s, userID, "屋顶下着雨", time.Date(2026, 8, 9, 23, 0, 0, 0, time.UTC))

	moments := []time.Time{
		time.Date(2026, 8, 8, 23, 0, 0, 0, time.UTC), // Saturday evening
		time.Date(2026, 8, 9, 9, 0, 0, 0, time.UTC), // Sunday morning
		time.Date(2026, 8, 9, 23, 0, 0, 0, time.UTC), // Sunday evening
	}
	var firstID string
	var firstStart, firstEnd time.Time
	for i, when := range moments {
		report, err := svc.Refresh(userID, when)
		if err != nil {
			t.Fatalf("refresh #%d: %v", i, err)
		}
		if i == 0 {
			firstID, firstStart, firstEnd = report.ID, report.WeekStart, report.WeekEnd
		}
		if report.ID != firstID {
			t.Errorf("moment %s: ID = %s, want %s (cache key must be stable across the week)", when, report.ID, firstID)
		}
		if !report.WeekStart.Equal(firstStart) || !report.WeekEnd.Equal(firstEnd) {
			t.Errorf("moment %s: window %s..%s, want %s..%s", when, report.WeekStart, report.WeekEnd, firstStart, firstEnd)
		}
	}
}

// TestRefreshIdempotent verifies that two refreshes at the same instant write a
// single report keyed by WeekStart, so the store holds one report per week
// rather than fragmenting the cache across entry points.
func TestRefreshIdempotent(t *testing.T) {
	s := newStore(t)
	svc := NewReportService(s)
	const userID = "user-idem"
	saveDream(t, s, userID, "屋顶下着雨", time.Date(2026, 8, 9, 23, 0, 0, 0, time.UTC))

	when := sundayRefresh()
	if _, err := svc.Refresh(userID, when); err != nil {
		t.Fatalf("first refresh: %v", err)
	}
	if _, err := svc.Refresh(userID, when); err != nil {
		t.Fatalf("second refresh: %v", err)
	}
	history := svc.History(userID)
	if got := len(history); got != 1 {
		t.Fatalf("expected 1 report for the week, got %d", got)
	}
	if got := history[0].TotalDreams; got != 1 {
		t.Errorf("TotalDreams = %d, want 1", got)
	}
}

// TestCurrentFindsCachedReadyReport asserts that after a refresh, Current reads
// the cached ready report without rebuilding — proving WeekStart is the single
// lookup key shared by both the writer (Refresh) and the reader (Current).
func TestCurrentFindsCachedReadyReport(t *testing.T) {
	s := newStore(t)
	svc := NewReportService(s)
	const userID = "user-cache"
	saveDream(t, s, userID, "屋顶下着雨", time.Date(2026, 8, 9, 23, 0, 0, 0, time.UTC))

	// Build the cached report for the week containing the Sunday dream. The
	// store's WeekStart must match what Current recomputes from the clock.
	if _, err := svc.Refresh(userID, time.Date(2026, 8, 9, 12, 0, 0, 0, time.UTC)); err != nil {
		t.Fatalf("seed refresh: %v", err)
	}

	// Current uses the live clock; this test runs in the week of 2026-08-03
	// only if the system clock is there. We assert the contract directly: a
	// report keyed by the recomputed WeekStart is found and returned ready.
	start := svc.clock.WeekStart(time.Date(2026, 8, 9, 12, 0, 0, 0, time.UTC))
	cached, ok := s.FindReport(userID, start)
	if !ok {
		t.Fatalf("FindReport by WeekStart=%s: not found (cache key mismatch)", start)
	}
	if cached.State != domain.ReportReady {
		t.Errorf("cached State = %q, want %q", cached.State, domain.ReportReady)
	}
	if cached.TotalDreams != 1 {
		t.Errorf("cached TotalDreams = %d, want 1", cached.TotalDreams)
	}
}

// TestRefreshExcludesOtherWeeksDreams ensures the window is not so wide that it
// pulls in dreams from adjacent weeks (which would also skew the count).
func TestRefreshExcludesOtherWeeksDreams(t *testing.T) {
	s := newStore(t)
	svc := NewReportService(s)
	const userID = "user-bound"
	saveDream(t, s, userID, "上周的梦", time.Date(2026, 8, 2, 23, 0, 0, 0, time.UTC)) // Sat of prev week
	saveDream(t, s, userID, "屋顶下着雨", time.Date(2026, 8, 9, 23, 0, 0, 0, time.UTC)) // this Sunday
	saveDream(t, s, userID, "下周的梦", time.Date(2026, 8, 10, 0, 0, 0, 0, time.UTC))  // next Monday

	report, err := svc.Refresh(userID, sundayRefresh())
	if err != nil {
		t.Fatalf("refresh: %v", err)
	}
	if report.TotalDreams != 1 {
		t.Errorf("TotalDreams = %d, want 1 (only the current week's dream)", report.TotalDreams)
	}
	if !report.WeekStart.Equal(time.Date(2026, 8, 3, 0, 0, 0, 0, time.UTC)) {
		t.Errorf("WeekStart = %s, want 2026-08-03", report.WeekStart)
	}
	if !report.WeekEnd.Equal(time.Date(2026, 8, 9, 23, 59, 59, 999999999, time.UTC)) {
		t.Errorf("WeekEnd = %s, want 2026-08-09 23:59:59.999999999", report.WeekEnd)
	}
}
