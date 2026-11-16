package subscriber

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/ship-monitor/cloud/pkg/connector/asyncapi"
	"go.uber.org/zap"
)

type Info interface {
	ID() uuid.UUID
	CheckSecurityKey(key string) bool
}

type Publisher interface {
	PublishConnectResponse(ctx context.Context, response *asyncapi.ConnectResponseSchema) error
}

type S struct {
	controller *asyncapi.UserController
	info       Info
	logger     *zap.Logger
	pub        Publisher

	cancel context.CancelFunc
}

func New(
	controller *asyncapi.UserController,
	info Info,
	logger *zap.Logger,
	pub Publisher,
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
	ctx, cancel := context.WithCancel(ctx)
	s.cancel = cancel

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
	approve := s.info.CheckSecurityKey(message.Payload.SecretKey)

	s.logger.Info("Connect request received",
		zap.String("requestID", message.Payload.RequestID.String()),
		zap.Bool("approve", approve),
	)

	err := s.pub.PublishConnectResponse(ctx,
		&asyncapi.ConnectResponseSchema{
			Approve:   approve,
			RequestID: message.Payload.RequestID,
		},
	)
	if err != nil {
		s.logger.Error("Publish connect response", zap.Error(err))

		return fmt.Errorf("publish connect response: %w", err)
	}

	return nil
}
