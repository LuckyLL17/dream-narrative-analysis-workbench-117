package store

import (
	"dream117/internal/domain"
	"dream117/pkg/collections"
)

func (
	s *Store,
) ReportDigests(
	userID string,
) []domain.ReportDigest {
	s.mu.RLock()
	items := make(
		[]domain.WeeklyReport, 0)
	for i := range s.data.Reports {
		report := s.data.Reports[i]
		if report.UserID == userID {
			items = append(
				items,
				report,
			)
		}
	}
	s.mu.RUnlock()
	collections.SortBy(
		items,
		newestReportFirst,
	)
	digests := make(
		[]domain.ReportDigest,
		0,
		len(items),
	)
	for i := range items {
		report := items[i]
		theme := ""
		if len(report.TopThemes) >
			0 {
			theme = report.TopThemes[0].Name
		}
		emotion := dominantEmotion(report.EmotionDist)
		digests = append(digests,
			domain.ReportDigest{WeekStart: report.WeekStart, WeekEnd: report.WeekEnd, Dreams: report.TotalDreams, Theme: theme, Emotion: emotion, Change: report.Summary, State: report.State})
	}
	return digests
}

func dominantEmotion(
	counts map[string]int,
) string {
	name, count := "", 0
	for e, n := range counts {
		if n > count {
			name, count = e, n
		}
	}
	return name
}
