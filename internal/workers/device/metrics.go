package device

import (
	"context"
	"fmt"
	"time"

	"github.com/ship-monitor/cloud/pkg/connector/asyncapi"
	"github.com/ship-monitor/hub-service/pkg/metrics"
	"go.uber.org/zap"
)

func (w *Worker) startPublishingMetrics() {
	ticker := time.NewTicker(w.config.SendMetricsTimeout)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			ctx, cancel := context.WithTimeout(context.Background(), ExecutionTimeout)
			defer cancel()

			err := w.publishMetrics(ctx)
			if err != nil {
				w.logger.Error("Publish device metrics", zap.Error(err))
			} else {
				w.logger.Info("Published device metrics")
			}
		case <-w.stop:
			w.logger.Info("Receive stop")

			return
		}
	}
}

func (w *Worker) publishMetrics(ctx context.Context) error {
	timestamp := time.Now().UTC()

	err := w.metric.PublishDeviceMetrics(ctx,
		[]asyncapi.MetricSchema{
			{
				Name:      metrics.Power,
				Timestamp: timestamp,
				Type:      asyncapi.TypePropertyFromMetricSchemaGAUGE,
				Unit:      "1",
				Value:     1,
			},
			{
				Name:      metrics.BatteryLevel,
				Timestamp: timestamp,
				Type:      asyncapi.TypePropertyFromMetricSchemaGAUGE,
				Unit:      "%",
				Value:     100, //nolint:mnd
			},
		})
	if err != nil {
		return fmt.Errorf("publish: %w", err)
	}

	return nil
}
