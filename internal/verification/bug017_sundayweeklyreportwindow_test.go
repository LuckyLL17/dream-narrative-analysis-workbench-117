package verification

import (
	"dream117/internal/analysis"
	"dream117/internal/domain"
	"dream117/internal/store"
	"dream117/pkg/clock"
	"testing"
	"time"
)

func TestBug017Sundayweeklyreportwindow(t *testing.T) {
	c := clock.Clock{}
	sunday := time.Date(2026, 8, 9, 15, 4, 5, 0, time.UTC)
	start, end := c.WeekStart(sunday), c.WeekEnd(sunday)
	wantStart := time.Date(2026, 8, 3, 0, 0, 0, 0, time.UTC)
	wantEnd := time.Date(2026, 8, 9, 23, 59, 59, 999999999, time.UTC)
	if !start.Equal(wantStart) || !end.Equal(wantEnd) {
		t.Fatalf("week window wrong: start=%s end=%s", start, end)
	}
	item := domain.Dream{ID: "sun", DreamDate: time.Date(2026, 8, 9, 23, 0, 0, 0, time.UTC), Emotion: domain.EmotionCalm, Clarity: 7, SleepHours: 8}
	report := analysis.BuildReport("u", []domain.Dream{item}, start, end, sunday)
	if report.TotalDreams != 1 || report.WeekStart != wantStart || report.WeekEnd != wantEnd {
		t.Fatalf("report lost Sunday record: %+v", report)
	}
}

func TestBug017RegressionHealth(t *testing.T) {
	s, err := store.Open(t.TempDir() + "/dreams.json")
	if err != nil {
		t.Fatal(err)
	}
	if s == nil {
		t.Fatal("store nil")
	}
}
