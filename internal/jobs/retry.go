package jobs

import (
	"time"

	"dream117/internal/domain"
	"dream117/internal/store"
)

func RetryFailed(
	data *store.Store,
) {
	snapshot := data.Snapshot()
	for i := range snapshot.Jobs {
		job := snapshot.Jobs[i]
		if job.Status == domain.AnalysisFailed && job.Attempts < 3 {
			_ = data.UpdateJob(job.ID, func(value *domain.AnalysisJob) {
				value.Status = domain.AnalysisQueued
				value.ScheduledAt = time.Now().UTC().Add(time.Duration(value.Attempts) * time.Second)
			})
		}
	}
}
