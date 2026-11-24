package sensors

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"github.com/ship-monitor/cloud/pkg/connector/asyncapi"
	"github.com/ship-monitor/hub-service/internal/types"
	"github.com/ship-monitor/hub-service/pkg/metrics"
	"go.uber.org/zap"
)

type Repo interface {
	GetSensors(ctx context.Context) ([]types.Sensor, error)
}

type Pub interface {
	PublishSensorMetrics(ctx context.Context, metrics asyncapi.SensorMetricsMessagePayload) error
}

// SendMetricsTimeout represents timeout between sending metrics.
const SendMetricsTimeout = time.Second * 30

const SendDeclarationTimeout = time.Minute * 2

type DeviceInfoService interface {
	ID() uuid.UUID
}

type Worker struct {
	controller *asyncapi.UserController
	logger     *zap.Logger
	info       DeviceInfoService
	pub        Pub
	repo       Repo
	stop       chan struct{}
}

func New(
	controller *asyncapi.UserController,
	logger *zap.Logger,
	info DeviceInfoService,
	pub Pub,
	repo Repo,
) *Worker {
	return &Worker{
		controller: controller,
		logger:     logger.Named("Sensors"),
		info:       info,
		pub:        pub,
		repo:       repo,
		stop:       make(chan struct{}),
	}
}

func (w *Worker) Start() {
	go w.startPublishingMetrics()
	go w.startPublishingDeclarations()
}

func (w *Worker) Stop() {
	w.stop <- struct{}{}

	w.stop <- struct{}{}
}

func (w *Worker) declare(ctx context.Context) error {
	s, err := w.repo.GetSensors(ctx)
	if err != nil {
		return fmt.Errorf("get sensors: %w", err)
	}

	err = w.controller.SendToSensorDeclarationOperation(
		ctx,
		asyncapi.SensorDeclarationsChannelParameters{
			DeviceId: w.info.ID().String(),
		},
		asyncapi.SensorDeclarationMessage{
			Headers: asyncapi.IdentificationHeadersSchema{
				Authorization: w.info.ID().String(),
			},
			Payload: asyncapi.SensorDeclarationSchema{
				Sensors: mapSensors(s),
			},
		},
	)
	if err != nil {
		return fmt.Errorf("send declaration: %w", err)
	}

	return nil
}
func (w *Worker) startPublishingDeclarations() {
	w.publishDeclaration()

	ticker := time.NewTicker(SendDeclarationTimeout)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			w.publishDeclaration()
		case <-w.stop:
			w.logger.Info("Receive stop")

			return
		}
	}
}

func (w *Worker) publishDeclaration() {
	ctx, cancel := context.WithTimeout(context.Background(), SendDeclarationTimeout)
	defer cancel()

	err := w.declare(ctx)
	if err != nil {
		w.logger.Error("Publish declaration", zap.Error(err))
	} else {
		w.logger.Info("Published declaration")
	}
}

func (w *Worker) startPublishingMetrics() {
	ticker := time.NewTicker(SendMetricsTimeout)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			ctx, cancel := context.WithTimeout(context.Background(), SendMetricsTimeout)
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

// temperature имитирует температуру холодильной камеры около -16.5°C.
//
//nolint:mnd, gosec
func (w *Worker) temperature() float64 {
	base := -16.5 + math.Sin(float64(time.Now().UnixNano())/1e9/60)*0.6

	return math.Round((base+rand.NormFloat64()*0.2)*10) / 10
}

func (w *Worker) publishMetric(ctx context.Context) error {
	sensors, err := w.repo.GetSensors(ctx)
	if err != nil {
		return fmt.Errorf("get sensors: %w", err)
	}

	timestamp := time.Now().UTC()

	for _, s := range sensors {
		value := w.temperature()

		err := w.pub.PublishSensorMetrics(ctx, asyncapi.SensorMetricsMessagePayload{
			SensorID: s.ID,
			Metrics: []asyncapi.MetricSchema{
				{
					Name:      metrics.Temperature,
					Type:      asyncapi.TypePropertyFromMetricSchemaGAUGE,
					Unit:      "C",
					Value:     value,
					Timestamp: timestamp,
				},
				{
					Name:      metrics.BatteryLevel,
					Type:      asyncapi.TypePropertyFromMetricSchemaGAUGE,
					Unit:      "%",
					Value:     100, //nolint:mnd
					Timestamp: timestamp,
				},
				{
					Name:      metrics.Power,
					Type:      asyncapi.TypePropertyFromMetricSchemaGAUGE,
					Unit:      "1",
					Value:     metrics.PowerConnected,
					Timestamp: timestamp,
				},
			},
		})
		if err != nil {
			return fmt.Errorf("publish: %w", err)
		}
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
