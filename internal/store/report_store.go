package store

import (
	"time"

	"dream117/internal/domain"
	"dream117/pkg/collections"
)

func (
	s *Store,
) SaveReport(
	report domain.WeeklyReport,
) error {
	return s.Update(func(data *domain.Database) error { data.Reports[report.ID] = report; return nil })
}

func (
	s *Store,
) FindReport(
	userID string,
	start time.Time,
) (domain.WeeklyReport, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for i := range s.data.Reports {
		report := s.data.Reports[i]
		if report.UserID == userID && report.WeekStart.Equal(start) {
			return report, true
		}
	}
	return domain.WeeklyReport{}, false
}

func (
	s *Store,
) ListReports(
	userID string,
) []domain.WeeklyReport {
	s.mu.RLock()
	defer s.mu.RUnlock()
	items := collections.FilterMap(s.data.Reports, func(_ string, report domain.WeeklyReport) bool { return report.UserID == userID })
	collections.SortBy(
		items,
		newestReportFirst,
	)
	return items
}

func (
	s *Store,
) MarkReportsPending(
	userID string,
) error {
	return s.Update(func(data *domain.Database) error {
		for id, report := range data.Reports {
			if report.UserID == userID {
				report.State = domain.ReportPending
				report.UpdatedAt = time.Now().UTC()
				data.Reports[id] = report
			}
		}
		return nil
	})
}
