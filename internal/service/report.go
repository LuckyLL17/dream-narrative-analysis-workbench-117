package service

import (
	"time"

	"dream117/internal/analysis"
	"dream117/internal/domain"
	"dream117/internal/store"
	"dream117/pkg/clock"
)

type ReportService struct {
	store *store.Store
	clock clock.Clock
}

func NewReportService(
	data *store.Store,
) *ReportService {
	return &ReportService{store: data, clock: clock.Clock{}}
}

func (
	s *ReportService,
) Refresh(
	userID string,
	when time.Time,
) (domain.WeeklyReport, error) {
	start := s.clock.WeekStart(when)
	end := s.clock.WeekEnd(when)
	items := s.store.ListDreams(userID, start, end, "")
	report := analysis.BuildReport(userID, items, start, end, time.Now().UTC())
	if err := s.store.SaveReport(report); err != nil {
		return domain.WeeklyReport{}, err
	}
	return report, nil
}
func (
	s *ReportService,
) Current(
	userID string,
) (domain.WeeklyReport, error) {
	start := s.clock.WeekStart(time.Now().UTC())
	if report, ok := s.store.FindReport(userID, start); ok && report.State == domain.ReportReady {
		return report, nil
	}
	return s.Refresh(userID, time.Now().UTC())
}
func (
	s *ReportService,
) History(
	userID string,
) []domain.WeeklyReport {
	return s.store.ListReports(userID)
}

func (
	s *ReportService,
) HistoryDigests(
	userID string,
) []domain.ReportDigest {
	return s.store.ReportDigests(userID)
}
