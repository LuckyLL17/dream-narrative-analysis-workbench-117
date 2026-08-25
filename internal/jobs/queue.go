package jobs

import (
	"time"

	"dream117/internal/domain"
	"dream117/internal/store"
	"dream117/pkg/ids"
)

type Queue struct{ store *store.Store }

func NewQueue(
	data *store.Store,
) *Queue {
	return &Queue{store: data}
}
func (
	q *Queue,
) Enqueue(
	userID,
	kind string,
) error {
	return q.store.EnqueueJob(domain.AnalysisJob{ID: ids.New("job"), UserID: userID, Kind: kind, Status: domain.AnalysisQueued, ScheduledAt: time.Now().UTC()})
}
func (
	q *Queue,
) Claim(
	limit int,
) []domain.AnalysisJob {
	jobs := q.store.PendingJobs(limit)
	for i := range jobs {
		job := jobs[i]
		_ = q.store.UpdateJob(
			job.ID, func(value *domain.AnalysisJob) { value.Status = domain.AnalysisRunning; value.Attempts++ })
	}
	return jobs
}
