package store

import (
	"time"

	"dream117/internal/domain"
	"dream117/pkg/collections"
)

func (
	s *Store,
) EnqueueJob(
	job domain.AnalysisJob,
) error {
	return s.Update(func(data *domain.Database) error { data.Jobs[job.ID] = job; return nil })
}

func (
	s *Store,
) PendingJobs(
	limit int,
) []domain.AnalysisJob {
	s.mu.RLock()
	defer s.mu.RUnlock()
	now := time.Now().UTC()
	items := collections.FilterMap(s.data.Jobs, func(_ string, job domain.AnalysisJob) bool {
		return job.Status == domain.AnalysisQueued && !job.ScheduledAt.After(now)
	})
	collections.SortBy(
		items,
		jobDueFirst,
	)
	if limit > 0 &&
		len(items) > limit {
		items = items[:limit]
	}
	return items
}

func (
	s *Store,
) UpdateJob(
	id string,
	update func(*domain.AnalysisJob),
) error {
	return s.Update(func(data *domain.Database) error {
		job, ok := data.Jobs[id]
		if !ok {
			return domain.ErrNotFound
		}
		update(&job)
		data.Jobs[id] = job
		return nil
	})
}
