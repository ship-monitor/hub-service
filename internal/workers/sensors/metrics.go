package sensors

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/ship-monitor/cloud/pkg/connector/asyncapi"
	"github.com/ship-monitor/hub-service/internal/types"
	"github.com/ship-monitor/hub-service/pkg/metrics"
	"github.com/ship-monitor/hub-service/pkg/sensor/driver"
	"go.uber.org/zap"
)

func (w *Worker) startPublishingMetrics() {
	ticker := time.NewTicker(w.config.SendMetricsTimeout)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			ctx, cancel := context.WithTimeout(context.Background(), w.config.SendMetricsTimeout)
			defer cancel()

			err := w.publishMetric(ctx)
			if err != nil {
				w.logger.Error("Publish sensor metrics", zap.Error(err))
			} else {
				w.logger.Info("Published sensor metrics")
			}
		case <-w.stop:
			w.logger.Info("Receive stop")

			return
		}
	}
}

func (w *Worker) publishMetric(ctx context.Context) error {
	sensors, err := w.repo.GetSensors(ctx)
	if err != nil {
		return fmt.Errorf("get sensors: %w", err)
	}

	for _, s := range sensors {
		err = errors.Join(w.publishSensorMetric(ctx, s.ID))
	}

	if err != nil {
		return fmt.Errorf("publish sensor metrics: %w", err)
	}

	return nil
}

//nolint:gochecknoglobals
var sensorMetrics = []struct {
	Name string
	Unit string
	Type asyncapi.TypePropertyFromMetricSchema
}{
	{
		Name: metrics.BatteryLevel,
		Unit: "%",
		Type: asyncapi.TypePropertyFromMetricSchemaGAUGE,
	},
	{
		Name: metrics.Power,
		Unit: "1",
		Type: asyncapi.TypePropertyFromMetricSchemaGAUGE,
	},
	{
		Name: metrics.Temperature,
		Unit: "C",
		Type: asyncapi.TypePropertyFromMetricSchemaGAUGE,
	},
}

var errReadMetrics = errors.New("read sensor metrics from driver")

func (w *Worker) publishSensorMetric(ctx context.Context, sensorID uuid.UUID) error {
	sensor, err := driver.New(sensorID)
	if err != nil {
		return fmt.Errorf("init driver: %w", err)
	}

	items := make([]asyncapi.MetricSchema, 0, len(sensorMetrics))

	for _, m := range sensorMetrics {
		metric, err := sensor.Read(ctx, m.Name)
		if err != nil {
			w.logger.Error("Read metric from driver", zap.Error(err))

			continue
		}

		items = append(items, asyncapi.MetricSchema{
			Name:      m.Name,
			Timestamp: metric.Timestamp,
			Type:      m.Type,
			Unit:      m.Unit,
			Value:     metric.Value,
		})
	}

	if len(items) == 0 {
		return errReadMetrics
	}

	err = w.pub.PublishSensorMetrics(ctx, asyncapi.SensorMetricsMessagePayload{
		SensorID: sensorID,
		Metrics:  items,
	})
	if err != nil {
		return fmt.Errorf("publish metric: %w", err)
	}

	return nil
}

func mapSensors(sensors []types.Sensor) []asyncapi.SensorDeclarationItemSchema {
	result := make([]asyncapi.SensorDeclarationItemSchema, 0, len(sensors))

	for _, s := range sensors {
		result = append(result, asyncapi.SensorDeclarationItemSchema{
			Id:   s.ID,
			Name: s.Name,
			Firmware: asyncapi.FirmwareSchema{
				Manufacturer: s.Manufacturer,
				Model:        s.Model,
				Revision:     s.Revision,
				Software:     s.Software,
			},
		})
	}

	return result
}
