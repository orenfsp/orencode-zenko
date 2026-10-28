package deadline

import (
	"context"
	"time"

	"github.com/orenfsp/orencode-zenko.git/pkg/logger"
	"go.uber.org/zap"
)

// RunReminderWorker periodically creates one reminder per deadline, recipient
// and lead time. The database unique key makes repeated worker runs safe.
func RunReminderWorker(ctx context.Context, repository Repository, interval time.Duration, daysBefore []int32) {
	if interval <= 0 || len(daysBefore) == 0 {
		return
	}
	run := func() {
		for _, days := range daysBefore {
			if days <= 0 {
				continue
			}
			if err := repository.CreateReportDeadlineReminders(ctx, days); err != nil && logger.Log != nil {
				logger.Log.Error("deadline reminder failed", zap.Int32("days_before", days), zap.Error(err))
			}
		}
	}
	run()
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			run()
		}
	}
}
