package publishers

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/ship-monitor/cloud/pkg/connector/asyncapi"
	"go.uber.org/zap"
)

type Info interface {
	ID() uuid.UUID
}
type Auth interface {
	GetAuthHeader(ctx context.Context) (string, error)
}

type P struct {
	controller *asyncapi.UserController
	logger     *zap.Logger
	info       Info
	auth       Auth
}

func New(
	controller *asyncapi.UserController,
	logger *zap.Logger,
	info Info,
	auth Auth,
) *P {
	return &P{
		controller: controller,
		logger:     logger,
		info:       info,
		auth:       auth,
	}
}

func (p *P) PublishDeviceMetrics(ctx context.Context, metrics []asyncapi.MetricSchema) error {
	header, err := p.auth.GetAuthHeader(ctx)
	if err != nil {
		return fmt.Errorf("unauthenticated: get auth header: %w", err)
	}

	err = p.controller.SendToDeviceMetricsOperation(ctx,
		asyncapi.DeviceMetricsChannelParameters{
			DeviceId: p.info.ID().String(),
		},
		asyncapi.MetricsMessage{
			Headers: asyncapi.IdentificationHeadersSchema{
				Authorization: header,
			},
			Payload: asyncapi.MetricsMessagePayload{
				Metrics: metrics,
			},
		})
	if err != nil {
		return fmt.Errorf("send metric: %w", err)
	}

	return nil
}

func (p *P) PublishConnectResponse(ctx context.Context, response *asyncapi.ConnectResponseSchema) error {
	header, err := p.auth.GetAuthHeader(ctx)
	if err != nil {
		return fmt.Errorf("unauthenticated: get auth header: %w", err)
	}

	err = p.controller.SendToOnConnectResponseOperation(
		ctx,
		asyncapi.ConnectResponsesChannelParameters{DeviceId: p.info.ID().String()},
		asyncapi.ConnectResponseMessage{
			Headers: asyncapi.IdentificationHeadersSchema{
				Authorization: header,
			},
			Payload: *response,
		},
	)
	if err != nil {
		return fmt.Errorf("send response: %w", err)
	}

	return nil
}
func (p *P) PublishHeartbeat(ctx context.Context, h *asyncapi.HeartbeatSchema) error {
	header, err := p.auth.GetAuthHeader(ctx)
	if err != nil {
		return fmt.Errorf("unauthenticated: get auth header: %w", err)
	}

	err = p.controller.SendToOnDeviceHeartbeatOperation(ctx,
		asyncapi.DevicesAliveChannelParameters{
			DeviceId: p.info.ID().String(),
		},
		asyncapi.HeartbeatMessage{
			Headers: asyncapi.IdentificationHeadersSchema{
				Authorization: header,
			},
			Payload: *h,
		},
	)
	if err != nil {
		return fmt.Errorf("send heartbeat: %w", err)
	}

	return nil
}

func (p *P) PublishSensorMetrics(ctx context.Context, metrics asyncapi.SensorMetricsMessagePayload) error {
	header, err := p.auth.GetAuthHeader(ctx)
	if err != nil {
		return fmt.Errorf("unauthenticated: get auth header: %w", err)
	}

	err = p.controller.SendToSensorMetricsOperation(
		ctx,
		asyncapi.SensorMetricsChannelParameters{
			DeviceId: p.info.ID().String(),
		},
		asyncapi.SensorMetricsMessage{
			Headers: asyncapi.IdentificationHeadersSchema{
				Authorization: header,
			},
			Payload: metrics,
		},
	)
	if err != nil {
		return fmt.Errorf("send sensor metrics: %w", err)
	}

	return nil
}
