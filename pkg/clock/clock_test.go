package clock

import (
	"testing"
	"time"
)

// weekEnd is the inclusive upper bound for a Monday-based week: Sunday
// 23:59:59.999999999 UTC. The Sunday of Aug 3-9 2026 is Aug 9.
var (
	augMonStart = time.Date(2026, 8, 3, 0, 0, 0, 0, time.UTC)
	augSunEnd   = time.Date(2026, 8, 9, 23, 59, 59, 999999999, time.UTC)
	augNextMon  = time.Date(2026, 8, 10, 0, 0, 0, 0, time.UTC)
	augNextSun  = time.Date(2026, 8, 16, 23, 59, 59, 999999999, time.UTC)
)

// weekStartCases exercises every weekday in the Aug 3-9 2026 week plus the two
// edge instants that used to be misclassified: Sunday at any time of day
// (rolled forward into next week) and the last nanosecond before midnight.
var weekStartCases = []struct {
	name  string
	value time.Time
}{
	{"Monday noon", time.Date(2026, 8, 3, 12, 0, 0, 0, time.UTC)},
	{"Tuesday noon", time.Date(2026, 8, 4, 12, 0, 0, 0, time.UTC)},
	{"Wednesday noon", time.Date(2026, 8, 5, 12, 0, 0, 0, time.UTC)},
	{"Thursday noon", time.Date(2026, 8, 6, 12, 0, 0, 0, time.UTC)},
	{"Friday noon", time.Date(2026, 8, 7, 12, 0, 0, 0, time.UTC)},
	{"Saturday noon", time.Date(2026, 8, 8, 12, 0, 0, 0, time.UTC)},
	{"Saturday last nanosecond", time.Date(2026, 8, 8, 23, 59, 59, 999999999, time.UTC)},
	{"Sunday midnight", time.Date(2026, 8, 9, 0, 0, 0, 0, time.UTC)},
	{"Sunday morning", time.Date(2026, 8, 9, 9, 30, 0, 0, time.UTC)},
	{"Sunday 23:00 dream time", time.Date(2026, 8, 9, 23, 0, 0, 0, time.UTC)},
	{"Sunday last nanosecond", time.Date(2026, 8, 9, 23, 59, 59, 999999999, time.UTC)},
}

func TestWeekStartCoversCurrentWeek(t *testing.T) {
	c := Clock{}
	for _, tc := range weekStartCases {
		t.Run(tc.name, func(t *testing.T) {
			got := c.WeekStart(tc.value)
			if !got.Equal(augMonStart) {
				t.Fatalf("WeekStart(%s) = %s, want %s", tc.value, got, augMonStart)
			}
		})
	}
}

func TestWeekEndCoversSunday(t *testing.T) {
	c := Clock{}
	for _, tc := range weekStartCases {
		t.Run(tc.name, func(t *testing.T) {
			got := c.WeekEnd(tc.value)
			if !got.Equal(augSunEnd) {
				t.Fatalf("WeekEnd(%s) = %s, want %s", tc.value, got, augSunEnd)
			}
		})
	}
}

// TestNextWeekStartsOnMonday ensures Sunday belongs to its own week, not the
// next one — the regression sent Sunday into the following week.
func TestNextWeekStartsOnMonday(t *testing.T) {
	c := Clock{}
	got := c.WeekStart(augNextMon)
	if !got.Equal(augNextMon) {
		t.Fatalf("WeekStart(next Monday) = %s, want %s", got, augNextMon)
	}
	gotEnd := c.WeekEnd(augNextMon)
	if !gotEnd.Equal(augNextSun) {
		t.Fatalf("WeekEnd(next Monday) = %s, want %s", gotEnd, augNextSun)
	}
}

// TestSaturdayAndSundayShareWindow is the cross-boundary consistency check:
// a refresh fired Saturday evening and one fired Sunday evening must resolve to
// the same week edges. Before the fix the Sunday refresh jumped a week ahead.
func TestSaturdayAndSundayShareWindow(t *testing.T) {
	c := Clock{}
	saturdayEvening := time.Date(2026, 8, 8, 23, 0, 0, 0, time.UTC)
	sundayEvening := time.Date(2026, 8, 9, 23, 0, 0, 0, time.UTC)
	satStart, satEnd := c.WeekStart(saturdayEvening), c.WeekEnd(saturdayEvening)
	sunStart, sunEnd := c.WeekStart(sundayEvening), c.WeekEnd(sundayEvening)
	if !satStart.Equal(sunStart) || !satEnd.Equal(sunEnd) {
		t.Fatalf("Saturday window %s..%s differs from Sunday window %s..%s",
			satStart, satEnd, sunStart, sunEnd)
	}
}
