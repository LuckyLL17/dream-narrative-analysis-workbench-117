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
		_, err := r.reports.Refresh(
			job.UserID, time.Now().UTC().Round(time.Second))
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
