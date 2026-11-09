package connect

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/ship-monitor/cloud/pkg/connector/asyncapi"
)

type Info interface {
	ID() uuid.UUID
}

type P struct {
	logger     *slog.Logger
	controller *asyncapi.UserController
	info       Info
}

func New(
	logger *slog.Logger,
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

	return nil
}
