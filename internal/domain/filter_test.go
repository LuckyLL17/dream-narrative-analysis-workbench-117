package domain

import (
	"testing"
	"time"

	"dream117/pkg/clock"
)

// TestDreamFilterNormalizedDayBounds asserts the date contract that a bare
// from/to date is floored/ceiled to an inclusive calendar-day range while a
// zero bound stays an open bound (no filtering on that side).
func TestDreamFilterNormalizedDayBounds(t *testing.T) {
	t.Run("bare_dates_become_inclusive_day_range", func(t *testing.T) {
		from := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
		to := time.Date(2026, 8, 2, 0, 0, 0, 0, time.UTC)
		f := DreamFilter{From: from, To: to, Page: 1, PageSize: 2}.Normalized()
		if f.From != clock.DayStart(from) {
			t.Fatalf("from not floored: got %s want %s", f.From, clock.DayStart(from))
		}
		if f.To != clock.DayEnd(to) {
			t.Fatalf("to not ceiled: got %s want %s", f.To, clock.DayEnd(to))
		}
		// The very last nanosecond of the end day must be inside the range.
		lastNs := clock.DayEnd(to)
		if lastNs.Before(f.From) || lastNs.After(f.To) {
			t.Fatalf("end-of-day %s not within [%s, %s]", lastNs, f.From, f.To)
		}
	})

	t.Run("zero_bounds_are_left_open", func(t *testing.T) {
		f := DreamFilter{Page: 1, PageSize: 2}.Normalized()
		if !f.From.IsZero() {
			t.Fatalf("unset from should remain zero, got %s", f.From)
		}
		if !f.To.IsZero() {
			t.Fatalf("unset to should remain zero, got %s", f.To)
		}
	})
}

// TestDreamFilterNormalizedPaging asserts page/page_size clamping.
func TestDreamFilterNormalizedPaging(t *testing.T) {
	cases := map[string]struct {
		in       DreamFilter
		page     int
		pageSize int
		minClear int
		maxSleep float64
	}{
		"defaults":         {DreamFilter{Page: 0, PageSize: 0}, 1, 24, 0, 0},
		"page_floor":       {DreamFilter{Page: -3, PageSize: 5}, 1, 5, 0, 0},
		"page_size_floor":  {DreamFilter{Page: 2, PageSize: 0}, 2, 24, 0, 0},
		"page_size_cap":    {DreamFilter{Page: 1, PageSize: 999}, 1, 100, 0, 0},
		"clarity_clamp":    {DreamFilter{Page: 1, PageSize: 5, MinimumClarity: 42}, 1, 5, 10, 0},
		"clarity_negative": {DreamFilter{Page: 1, PageSize: 5, MinimumClarity: -2}, 1, 5, 0, 0},
		"sleep_negative":   {DreamFilter{Page: 1, PageSize: 5, MaximumSleep: -1}, 1, 5, 0, 0},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := tc.in.Normalized()
			if got.Page != tc.page {
				t.Fatalf("page: got %d want %d", got.Page, tc.page)
			}
			if got.PageSize != tc.pageSize {
				t.Fatalf("page_size: got %d want %d", got.PageSize, tc.pageSize)
			}
			if got.MinimumClarity != tc.minClear {
				t.Fatalf("min_clarity: got %d want %d", got.MinimumClarity, tc.minClear)
			}
			if got.MaximumSleep != tc.maxSleep {
				t.Fatalf("max_sleep: got %v want %v", got.MaximumSleep, tc.maxSleep)
			}
		})
	}
}
