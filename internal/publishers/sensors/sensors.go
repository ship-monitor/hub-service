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

//nolint:gochecknoglobals
var sensors = []asyncapi.SensorDeclarationItemSchema{
	{Id: asyncapi.SensorID(uuid.NewString()), Name: "coldstat-sensor-1"},
	{Id: asyncapi.SensorID(uuid.NewString()), Name: "coldstat-sensor-2"},
}

// SendTimeout represents timeout between sending metrics.
const SendTimeout = time.Second * 30

type DeviceInfoService interface {
	ID() uuid.UUID
}

type Publisher struct {
	controller *asyncapi.UserController
	logger     *zap.Logger
	info       DeviceInfoService
	stop       chan struct{}
}

func New(
	controller *asyncapi.UserController,
	logger *zap.Logger,
	info DeviceInfoService,
) *Publisher {
	return &Publisher{
		controller: controller,
		logger:     logger.Named("Sensors"),
		info:       info,
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
	p.stop <- struct{}{}
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
				Sensors: sensors,
			},
		},
	)
	if err != nil {
		return fmt.Errorf("send declaration: %w", err)
	}

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

	timestamp := time.Now().UTC()

	for _, s := range sensors {
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
					SensorID: s.Id,
					Metrics: []asyncapi.MetricSchema{
						{
							Name:      TemperatureMetric,
							Type:      asyncapi.TypePropertyFromMetricSchemaGAUGE,
							Unit:      "C",
							Value:     value,
							Timestamp: timestamp,
						},
						{
							Name:      BatteryMetric,
							Type:      asyncapi.TypePropertyFromMetricSchemaGAUGE,
							Unit:      "%",
							Value:     100, //nolint:mnd
							Timestamp: timestamp,
						},
					},
				},
			},
		)
		if err != nil {
			p.logger.Error("Publish metric", zap.Error(err))

			return
		}

		p.logger.Info("Published metrics",
			zap.String("sensorId", string(s.Id)),
		)
	}
}

const TemperatureMetric = "temperature"
const BatteryMetric = "device.battery"
