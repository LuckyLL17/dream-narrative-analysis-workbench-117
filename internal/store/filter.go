package store

import (
	"sort"
	"strings"
	"time"

	"dream117/internal/domain"
	"dream117/pkg/clock"
	"dream117/pkg/collections"
)

func (
	s *Store,
) SearchPage(
	userID string,
	filter domain.DreamFilter,
) domain.DreamPage {
	filter = filter.Normalized()
	s.mu.RLock()
	items := collections.FilterMap(s.data.Dreams, func(_ string, d domain.Dream) bool {
		return d.UserID == userID && matchesFilter(d, filter)
	})
	s.mu.RUnlock()
	sort.SliceStable(items,
		func(i, j int) bool {
			if items[i].DreamDate.Equal(items[j].DreamDate) {
				return items[i].CreatedAt.After(items[j].CreatedAt)
			}
			return items[i].DreamDate.After(items[j].DreamDate)
		})
	total := len(items)
	// Page count is the count of full-or-partial pages, rounded up; an empty
	// result still reports a single page so callers see "page 1 of 1" rather
	// than the contradictory "total=0 total_pages=0".
	pages := total / filter.PageSize
	if total%filter.PageSize != 0 {
		pages++
	}
	if pages == 0 {
		pages = 1
	}
	// Echo the requested page back even when it falls past the last page: an
	// out-of-range page returns an empty slice plus correct total/total_pages
	// rather than silently substituting another page's contents. start/end are
	// clamped against total so the slice never panics.
	start := (filter.Page - 1) * filter.PageSize
	if start < 0 || start > total {
		start = total
	}
	end := start + filter.PageSize
	if end > total {
		end = total
	}
	pageItems := append([]domain.Dream(nil), items[start:end]...)
	return domain.DreamPage{Items: pageItems, Page: filter.Page, PageSize: filter.PageSize, Total: total, TotalPages: pages, HasNext: filter.Page < pages, HasPrevious: filter.Page > 1}
}

func matchesFilter(
	d domain.Dream,
	filter domain.DreamFilter,
) bool {
	if !filter.From.IsZero() && d.DreamDate.Before(filter.From) {
		return false
	}
	if !filter.To.IsZero() && d.DreamDate.After(filter.To) {
		return false
	}
	if filter.Emotion != "" && d.Emotion != filter.Emotion {
		return false
	}
	if filter.MinimumClarity > 0 && d.Clarity < filter.MinimumClarity {
		return false
	}
	if filter.MaximumSleep > 0 && d.SleepHours > filter.MaximumSleep {
		return false
	}
	if filter.RememberedOnly && !d.RememberDetail {
		return false
	}
	if filter.Theme != "" && !d.HasTheme(filter.Theme) {
		return false
	}
	return phraseMatches(d, filter.Query)
}

func phraseMatches(
	d domain.Dream,
	query string,
) bool {
	query = strings.ToLower(
		strings.TrimSpace(query))
	if query == "" {
		return true
	}
	parts := []string{d.Title, d.Content, strings.Join(d.Tags, " ")}
	for i := range d.Themes {
		hit := d.Themes[i]
		parts = append(parts,
			hit.Name, strings.Join(hit.Evidence, " "))
	}
	return strings.Contains(strings.ToLower(strings.Join(parts, " ")), query)
}

func DateFloor(
	value time.Time,
) time.Time {
	return clock.DayStart(value)
}

func DateCeil(
	value time.Time,
) time.Time {
	return clock.DayEnd(value)
}

func NewFilter(
	from,
	to time.Time,
	query string,
) domain.DreamFilter {
	return domain.DreamFilter{From: DateFloor(from), To: DateCeil(to), Query: query, Page: 1, PageSize: 24}
}
