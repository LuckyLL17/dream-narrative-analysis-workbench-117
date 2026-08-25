package service

import (
	"time"

	"dream117/internal/store"
)

type HealthService struct {
	store   *store.Store
	started time.Time
}

func NewHealthService(
	data *store.Store,
) *HealthService {
	return &HealthService{store: data, started: time.Now().UTC()}
}
func (s *HealthService) Status() map[string]interface{} {
	snapshot := s.store.Snapshot()
	staleReports := 0
	for i := range snapshot.Reports {
		report := snapshot.Reports[i]
		if report.State != "ready" {
			staleReports++
		}
	}
	return map[string]interface{}{"status": "ok", "started_at": s.started, "updated_at": snapshot.UpdatedAt, "users": len(snapshot.Users), "dreams": len(snapshot.Dreams), "elements": len(snapshot.Elements), "reports": len(snapshot.Reports), "jobs": len(snapshot.Jobs), "stale_reports": staleReports, "schema": snapshot.Schema}
}
