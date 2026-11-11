package subscriber

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/ship-monitor/cloud/pkg/connector/asyncapi"
	"github.com/ship-monitor/hub-service/internal/publishers/connect"
	"go.uber.org/zap"
)

type Info interface {
	ID() uuid.UUID
}

type S struct {
	controller *asyncapi.UserController
	info       Info
	logger     *zap.Logger
	pub        *connect.P

	cancel context.CancelFunc
}

func New(
	controller *asyncapi.UserController,
	info Info,
	logger *zap.Logger,
	pub *connect.P,
) *S {
	return &S{
		controller: controller,
		info:       info,
		logger:     logger.Named("Subscriber"),
		pub:        pub,
		cancel:     nil,
	}
}
func (s *S) Start(ctx context.Context) error {
	err := s.controller.SubscribeToPublishConnectRequestOperation(
		ctx,
		asyncapi.ConnectRequestsChannelParameters{DeviceId: s.info.ID().String()},
		s.HandleConnectRequest,
	)
	if err != nil {
		return fmt.Errorf("subscribe: %w", err)
	}

	return nil
}
func (s *S) Stop() {
	s.cancel()
}

func (s *S) HandleConnectRequest(ctx context.Context, message asyncapi.ConnectRequestMessage) error {
	s.logger.Info("Connect request received", zap.String("requestID", message.Payload.RequestID.String()))

	err := s.pub.PublishConnectResponse(ctx, &asyncapi.ConnectResponseMessage{
		Payload: asyncapi.ConnectResponseSchema{
			Approve:   true,
			RequestID: message.Payload.RequestID,
		},
		Headers: asyncapi.IdentificationHeadersSchema{
			Authorization: s.info.ID().String(),
		},
	})
	if err != nil {
		s.logger.Error("Publish connect response", zap.Error(err))

		return fmt.Errorf("publish connect response: %w", err)
	}

	return nil
}
