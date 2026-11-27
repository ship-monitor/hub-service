package sensors

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/ship-monitor/cloud/pkg/connector/asyncapi"
	"github.com/ship-monitor/hub-service/pkg/sensor/driver"
	"go.uber.org/zap"
)

func (w *Worker) startCheckingTemp() {
	ctx, cancel := context.WithTimeout(context.Background(), w.config.CheckTemperatureTimeout)
	defer cancel()

	err := w.CheckTemperature(ctx)
	if err != nil {
		w.logger.Error("Check temperature on startup", zap.Error(err))
	}

	ticker := time.NewTicker(w.config.CheckTemperatureTimeout)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			ctx, cancel := context.WithTimeout(context.Background(), w.config.CheckTemperatureTimeout)
			defer cancel()

			err := w.CheckTemperature(ctx)
			if err != nil {
				w.logger.Error("Check temperature", zap.Error(err))
			} else {
				w.logger.Info("Temperature checked")
			}
		case <-w.stop:
			w.logger.Info("Receive stop")

			return
		}
	}
}

func (w *Worker) CheckTemperature(ctx context.Context) error {
	sensors, err := w.repo.GetSensors(ctx)
	if err != nil {
		return fmt.Errorf("get sensors: %w", err)
	}

	for _, s := range sensors {
		err = errors.Join(w.checkTemp(ctx, s.ID))
	}

	if err != nil {
		return fmt.Errorf("chek sensor temp: %w", err)
	}

	return nil
}

func (w *Worker) checkTemp(ctx context.Context, sensorID uuid.UUID) error {
	sensor, err := driver.New(sensorID)
	if err != nil {
		return fmt.Errorf("init driver: %w", err)
	}

	temp, err := sensor.ReadTemperature(ctx)
	if err != nil {
		return fmt.Errorf("read temperature: %w", err)
	}

	if temp.Value >= -16.9 {
		err := w.pub.PublishNotification(ctx, asyncapi.NotificationSchema{
			Type:       "temperature-high",
			OccurredAt: temp.Timestamp,
			ReportedAt: time.Now(),
			Source: asyncapi.SourcePropertyFromNotificationSchema{
				Id:   sensorID,
				Type: asyncapi.NotificationSourceSchemaSensor,
			},
			Severity: asyncapi.SeveritySchemaCritical,
			Data: &asyncapi.DataPropertyFromNotificationSchema{
				AdditionalProperties: map[string]any{
					"temperature": temp.Value,
				},
			},
		})
		if err != nil {
			return fmt.Errorf("publish notification: %w", err)
		}
	}

	if temp.Value <= -17 {
		err := w.pub.PublishNotification(ctx, asyncapi.NotificationSchema{
			Type:       "temperature-low",
			OccurredAt: temp.Timestamp,
			ReportedAt: time.Now(),
			Source: asyncapi.SourcePropertyFromNotificationSchema{
				Id:   sensorID,
				Type: asyncapi.NotificationSourceSchemaSensor,
			},
			Severity: asyncapi.SeveritySchemaCritical,
			Data: &asyncapi.DataPropertyFromNotificationSchema{
				AdditionalProperties: map[string]any{
					"temperature": temp.Value,
				},
			},
		})
		if err != nil {
			return fmt.Errorf("publish notification: %w", err)
		}
	}

	return nil
}
