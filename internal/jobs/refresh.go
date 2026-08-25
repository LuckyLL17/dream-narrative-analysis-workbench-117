package jobs

import (
	"log"
	"time"

	"dream117/internal/domain"
	"dream117/internal/service"
)

type Refresher struct {
	queue   *Queue
	reports *service.ReportService
}

func NewRefresher(
	queue *Queue,
	reports *service.ReportService,
) *Refresher {
	return &Refresher{queue: queue, reports: reports}
}
func (r *Refresher) RunOnce() {
	jobs := r.queue.Claim(8)
	for i := range jobs {
		job := jobs[i]
		// Use the raw instant: rounding a Sunday-even timestamp to the nearest
		// second can land on Monday 00:00:00 and shift the week edge, so the
		// background refresh would disagree with an HTTP refresh fired at the
		// same moment. The service shares this instant for the week edges and
		// the report timestamps.
		_, err := r.reports.Refresh(
			job.UserID, time.Now().UTC())
		_ = r.queue.store.UpdateJob(job.ID, func(value *domain.AnalysisJob) {
			value.Status = domain.AnalysisDone
			value.FinishedAt = ptr(
				time.Now().UTC())
			if err != nil {
				value.Status = domain.AnalysisFailed
				value.LastError = err.Error()
			}
		})
		if err != nil {
			log.Printf("analysis job failed id=%s err=%v", job.ID, err)
		}
	}
}
func ptr(
	value time.Time,
) *time.Time {
	return &value
}
