package analysis

import (
	"testing"
	"time"

	"dream117/internal/domain"
)

func dream(title string, dreamDate time.Time) domain.Dream {
	return domain.Dream{
		ID:        title,
		Title:     title,
		Content:   "醒来记得很清楚的一段叙事内容",
		DreamDate: dreamDate,
		Emotion:   domain.EmotionAfraid,
		Clarity:   7,
		SleepHours: 7.5,
	}
}

// TestBuildReportIncludesSundayDream is the report-construction half of the
// regression: given the correct Monday-to-Sunday edges, BuildReport must keep
// the Sunday-night dream 屋顶下着雨 in TotalDreams.
func TestBuildReportIncludesSundayDream(t *testing.T) {
	start := time.Date(2026, 8, 3, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 8, 9, 23, 59, 59, 999999999, time.UTC)
	when := time.Date(2026, 8, 9, 23, 0, 0, 0, time.UTC)

	items := []domain.Dream{
		dream("上周六的梦", time.Date(2026, 8, 2, 23, 0, 0, 0, time.UTC)), // outside (prev week)
		dream("周三的梦", time.Date(2026, 8, 5, 7, 0, 0, 0, time.UTC)),   // inside
		dream("屋顶下着雨", time.Date(2026, 8, 9, 23, 0, 0, 0, time.UTC)), // Sunday edge — must be inside
		dream("下周的梦", time.Date(2026, 8, 10, 0, 0, 0, 0, time.UTC)),   // outside (next week)
	}

	report := BuildReport("user-build", items, start, end, when)

	if report.WeekStart != start {
		t.Errorf("WeekStart = %s, want %s", report.WeekStart, start)
	}
	if report.WeekEnd != end {
		t.Errorf("WeekEnd = %s, want %s", report.WeekEnd, end)
	}
	if report.TotalDreams != 2 {
		t.Errorf("TotalDreams = %d, want 2 (Wednesday + Sunday only)", report.TotalDreams)
	}
	if report.State != domain.ReportReady {
		t.Errorf("State = %q, want %q", report.State, domain.ReportReady)
	}
	if report.CreatedAt != when {
		t.Errorf("CreatedAt = %s, want the refresh instant %s", report.CreatedAt, when)
	}
}

// TestBuildReportSundayEndBoundary locks the inclusive upper bound: a dream
// logged at the very last nanosecond of Sunday stays inside the report.
func TestBuildReportSundayEndBoundary(t *testing.T) {
	start := time.Date(2026, 8, 3, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 8, 9, 23, 59, 59, 999999999, time.UTC)
	when := start

	items := []domain.Dream{dream("last nanosecond", end)}
	report := BuildReport("user-edge", items, start, end, when)
	if report.TotalDreams != 1 {
		t.Errorf("TotalDreams = %d, want 1 (last-nanosecond dream is in-bounds)", report.TotalDreams)
	}
}

// TestBuildReportEmptyWindow checks the empty-week state so a future change to
// the filter predicate cannot silently drop to zero for a valid full week.
func TestBuildReportEmptyWindow(t *testing.T) {
	start := time.Date(2026, 8, 3, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 8, 9, 23, 59, 59, 999999999, time.UTC)
	when := start
	report := BuildReport("user-empty", nil, start, end, when)
	if report.TotalDreams != 0 {
		t.Errorf("TotalDreams = %d, want 0", report.TotalDreams)
	}
	if report.State != domain.ReportPending {
		t.Errorf("State = %q, want %q", report.State, domain.ReportPending)
	}
}
