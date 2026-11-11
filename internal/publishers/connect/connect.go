package connect

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

type P struct {
	logger     *zap.Logger
	controller *asyncapi.UserController
	info       Info
}

func New(
	logger *zap.Logger,
	controller *asyncapi.UserController,
	info Info,
) *P {
	return &P{
		logger:     logger,
		controller: controller,
		info:       info,
	}
}

func (p *P) PublishConnectResponse(ctx context.Context, response *asyncapi.ConnectResponseMessage) error {
	err := p.controller.SendToOnConnectResponseOperation(
		ctx,
		asyncapi.ConnectResponsesChannelParameters{DeviceId: p.info.ID().String()},
		*response,
	)
	if err != nil {
		return fmt.Errorf("send response: %w", err)
	}

	p.logger.Info("Connect response published")

	return nil
}
