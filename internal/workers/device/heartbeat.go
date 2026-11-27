package device

import (
	"context"
	"fmt"
	"time"

	"github.com/ship-monitor/cloud/pkg/connector/asyncapi"
	"go.uber.org/zap"
)

func (w *Worker) startPublishingHeartbeat() {
	ticker := time.NewTicker(w.config.SendHeartbeatTimeout)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			ctx, cancel := context.WithTimeout(context.Background(), ExecutionTimeout)
			defer cancel()

			err := w.publishHeartbeat(ctx)
			if err != nil {
				w.logger.Error("Publish heartbeat", zap.Error(err))
			} else {
				w.logger.Info("Published heartbeat")
			}
		case <-w.stop:
			w.logger.Info("Receive stop")

			return
		}
	}
}

func (w *Worker) publishHeartbeat(ctx context.Context) error {
	hb := NewHeartbeat()

	err := w.heartbeat.PublishHeartbeat(ctx, hb)
	if err != nil {
		return fmt.Errorf("publish: %w", err)
	}

	return nil
}

func NewHeartbeat() *asyncapi.HeartbeatSchema {
	return &asyncapi.HeartbeatSchema{
		Timestamp: time.Now().UTC(),
	}
}
