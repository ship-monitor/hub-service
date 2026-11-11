package sensors

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"github.com/ship-monitor/cloud/pkg/connector/asyncapi"
	"go.uber.org/zap"
)

// SendTimeout represents timeout between sending metrics.
const SendTimeout = time.Second * 30

type DeviceInfoService interface {
	ID() uuid.UUID
}

type Publisher struct {
	controller *asyncapi.UserController
	logger     *zap.Logger
	info       DeviceInfoService
	sensorID   asyncapi.SensorID
	sensorName string
	stop       chan struct{}
}

func New(
	controller *asyncapi.UserController,
	logger *zap.Logger,
	info DeviceInfoService,
	sensorID string,
	sensorName string,
) *Publisher {
	return &Publisher{
		controller: controller,
		logger:     logger,
		info:       info,
		sensorID:   asyncapi.SensorID(sensorID),
		sensorName: sensorName,
		stop:       make(chan struct{}),
	}
}

func (p *Publisher) Start() error {
	err := p.declare()
	if err != nil {
		return fmt.Errorf("declare sensors: %w", err)
	}

	go p.startPublishing()

	return nil
}

func (p *Publisher) Stop() {
	close(p.stop)
}

func (p *Publisher) declare() error {
	ctx := context.Background()

	err := p.controller.SendToSensorDeclarationOperation(
		ctx,
		asyncapi.SensorDeclarationsChannelParameters{
			DeviceId: p.info.ID().String(),
		},
		asyncapi.SensorDeclarationMessage{
			Headers: asyncapi.IdentificationHeadersSchema{
				Authorization: p.info.ID().String(),
			},
			Payload: asyncapi.SensorDeclarationSchema{
				Sensors: []asyncapi.SensorDeclarationItemSchema{
					{Id: p.sensorID, Name: p.sensorName},
				},
			},
		},
	)
	if err != nil {
		return fmt.Errorf("send declaration: %w", err)
	}

	p.logger.Info(
		"Sensors declared",
		zap.Object("sensor", zap.DictObject(
			zap.String("id", string(p.sensorID)),
			zap.String("name", p.sensorName),
		)),
	)

	return nil
}

func (p *Publisher) startPublishing() {
	ticker := time.NewTicker(SendTimeout)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			p.publishMetric()
		case <-p.stop:
			p.logger.Info("Receive stop")

			return
		}
	}
}

// temperature имитирует температуру холодильной камеры около -16.5°C.
//
//nolint:mnd, gosec
func (p *Publisher) temperature() float64 {
	base := -16.5 + math.Sin(float64(time.Now().UnixNano())/1e9/60)*0.6

	return math.Round((base+rand.NormFloat64()*0.2)*10) / 10
}

func (p *Publisher) publishMetric() {
	ctx := context.Background()

	value := p.temperature()

	err := p.controller.SendToMetricsOperationForSensorMetrics(
		ctx,
		asyncapi.MetricsChannelParameters{
			DeviceId: p.info.ID().String(),
		},
		asyncapi.SensorMetricsMessage{
			Headers: asyncapi.IdentificationHeadersSchema{
				Authorization: p.info.ID().String(),
			},
			Payload: asyncapi.SensorMetricsMessagePayload{
				SensorID: p.sensorID,
				Metrics: []asyncapi.MetricSchema{
					{
						Name:      "temperature",
						Type:      asyncapi.TypePropertyFromMetricSchemaGAUGE,
						Unit:      "C",
						Value:     value,
						Timestamp: time.Now().UTC(),
					},
				},
			},
		},
	)
	if err != nil {
		p.logger.Error("Publish metric", zap.Error(err))

		return
	}

	p.logger.Info("Published metric",
		zap.String("sensorId", string(p.sensorID)),
		zap.Float64("value", value),
	)
}
