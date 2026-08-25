package jobs

import (
	"context"
	"time"
)

type Scheduler struct {
	every     time.Duration
	refresher *Refresher
}

func NewScheduler(
	every time.Duration,
	refresher *Refresher,
) *Scheduler {
	return &Scheduler{every: every, refresher: refresher}
}
func (
	s *Scheduler,
) Start(
	ctx context.Context,
) {
	ticker := time.NewTicker(s.every)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				s.refresher.RunOnce()
			case <-ctx.Done():
				return
			}
		}
	}()
}
